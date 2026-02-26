import os
import pickle
import argparse

import torch
import numpy as np
import pandas as pd
import torch.nn as nn
from tqdm import tqdm
import torch.optim as optim
from sklearn.metrics import (
    r2_score,
    mean_squared_error,
    mean_absolute_error,
)
from sklearn.preprocessing import MinMaxScaler
from sklearn.model_selection import train_test_split

import onnxruntime as ort

# configuration
DATA_FILE = "housing_data.csv"
MODEL_ONNX_FILE = "house_price_model.onnx"
SCALER_FILE = "scalers.pkl"

if torch.backends.mps.is_available():
    DEVICE = torch.device("mps")
elif torch.cuda.is_available():
    DEVICE = torch.device("cuda")
else:
    DEVICE = torch.device("cpu")

print(f"using device: {DEVICE}")


class HousePricePredictor(nn.Module):
    def __init__(self) -> None:
        super(HousePricePredictor, self).__init__()

        # input layer (3 features) -> hidden layer 1 (64 neurons)
        self.fc1 = nn.Linear(3, 64)

        # hidden layer 1 (64 neurons) -> hidden layer 2 (32 neurons)
        self.fc2 = nn.Linear(64, 32)

        # hidden layer 2 (32 neurons) -> output layer (1 neuron)
        self.fc3 = nn.Linear(32, 1)

        # activation function
        self.relu = nn.ReLU()

    def forward(self, x):
        x = self.relu(self.fc1(x))
        x = self.relu(self.fc2(x))
        x = self.fc3(x)  # output layer, so no activation needed
        return x


def load_and_preprocess_data(
    data_path: str,
    test_size: float = 0.2,
    validation_size: float = 0.2,
    random_state: int = 42,
):
    try:
        df = pd.read_csv(data_path)
    except FileNotFoundError:
        print(f"error: data file not found at {data_path}")
        return None, None, None, None, None, None, None, None

    # define features (X) and target (y)
    features = ["square_footage", "bedrooms", "bathrooms"]
    target = "price_thousands"

    X = df[features].values
    y = df[[target]].values

    # initialize minmaxscaler for features and targets
    features_scaler = MinMaxScaler()
    target_scaler = MinMaxScaler()

    # fit and transform features and target
    X_scaled = features_scaler.fit_transform(X)
    y_scaled = target_scaler.fit_transform(y)

    # first split: separate out the test set
    X_train_val, X_test, y_train_val, y_test = train_test_split(
        X_scaled, y_scaled, test_size=test_size, random_state=random_state
    )

    # second split: split the remaining data into training and validation sets
    val_ratio_in_train_val = validation_size / (1 - test_size)
    X_train, X_val, y_train, y_val = train_test_split(
        X_train_val,
        y_train_val,
        test_size=val_ratio_in_train_val,
        random_state=random_state,
    )

    # convert to pytorch tensors and move to selected devices
    X_train_tensor = torch.tensor(X_train, dtype=torch.float32).to(DEVICE)
    y_train_tensor = torch.tensor(y_train, dtype=torch.float32).to(DEVICE)
    X_val_tensor = torch.tensor(X_val, dtype=torch.float32).to(DEVICE)
    y_val_tensor = torch.tensor(y_val, dtype=torch.float32).to(DEVICE)
    X_test_tensor = torch.tensor(X_test, dtype=torch.float32).to(DEVICE)
    y_test_tensor = torch.tensor(y_test, dtype=torch.float32).to(DEVICE)

    print(f"data loaded and preprocessed from {data_path}")
    print(
        f"training samples: {X_train_tensor.shape[0]}, "
        f"validation samples: {X_val_tensor.shape[0]}, "
        f"test samples: {X_test_tensor.shape[0]}"
    )

    return (
        X_train_tensor,
        y_train_tensor,
        X_val_tensor,
        y_val_tensor,
        X_test_tensor,
        y_test_tensor,
        features_scaler,
        target_scaler,
    )


def train_model(
    model,
    X_train,
    y_train,
    X_val,
    y_val,
    X_test,
    y_test,
    target_scaler,
    num_epochs=1000,
    learning_rate=0.001,
    patience=50,
    min_delta=0.0001,
):
    """
    Trains the neural network model with early stopping based on validation loss
    and reports evaluation metrics on the test set.
    Args:
        model (nn.Module): The neural network model to train.
        X_train (torch.Tensor): Training features.
        y_train (torch.Tensor): Training target.
        X_val (torch.Tensor): Validation features.
        y_val (torch.Tensor): Validation target.
        X_test (torch.Tensor): Test features.
        y_test (torch.Tensor): Test target.
        target_scaler (MinMaxScaler): The scaler used for the target variable, needed for inverse transform.
        num_epochs (int): Maximum number of training epochs.
        learning_rate (float): Learning rate for the optimizer.
        patience (int): Number of epochs to wait for improvement before stopping.
        min_delta (float): Minimum change in monitored quantity to qualify as an improvement.
    """

    # define loss function and optimizer
    criterion = nn.MSELoss()
    optimizer = optim.Adam(model.parameters(), lr=learning_rate)

    best_val_loss = float("inf")
    epochs_no_improve = 0
    early_stop = False
    best_model_state = None  # to store the state_dict of the best model

    print(
        f"\nStarting model training for {num_epochs} epochs with early stopping"
        f" (patience = {patience})..."
    )

    try:
        with tqdm(range(num_epochs), desc="Training Progress") as pbar:
            for epochs in pbar:
                if early_stop:
                    break

                # set model to training mode
                model.train()

                # execute foward pass (training)
                outputs = model(X_train)
                loss = criterion(outputs, y_train)

                # backwards pass and optimize
                optimizer.zero_grad()  # clear gradients
                loss.backward()  # computing gradients
                optimizer.step()

                # evaluate on validation set
                model.eval()  # set the model to evaluation mode

                with torch.no_grad():  # disabling gradient calculation
                    val_outputs = model(X_val)
                    val_loss = criterion(val_outputs, y_val)

                # early stopping logic based on validation loss
                if val_loss.item() < best_val_loss - min_delta:
                    best_val_loss = val_loss.item()
                    epochs_no_improve = 0
                    best_model_state = (
                        model.state_dict()
                    )  # save best model state
                else:
                    epochs_no_improve += 1
                    if epochs_no_improve == patience:
                        print(
                            f"\nEarly stopping triggered at epoch {epochs + 1} (no improvement for {patience} epochs)"
                        )
                        early_stop = True

                # update tqdm post-fix with current loss values
                pbar.set_postfix_str(
                    f"train loss: {loss.item():.4f}, val loss: {val_loss.item():.4f}"
                )

        print("training completed!")

        # load the best model state if early stopping occured and a best state was saved
        if best_model_state:
            model.load_state_dict(best_model_state)
            print("loaded best model state for final evaluation and saving")
        else:
            # if no improvement was ever found (e.g. patience=0 or very small min_delta)
            # the last state of the model is used
            print(
                "no improvement found during traininig using final model state for evaluation"
            )

        # final model evaluation on test set
        print("\n--- final model evaluation on test set ---")
        model.eval()

        with torch.no_grad():
            # get predictions on the test set
            test_predictions_scaled = model(X_test).cpu().numpy()
            y_test_cpu = y_test.cpu().numpy()

            # inverse transform predictions and actual values to original scale for interpretable metrics
            actual_prices = target_scaler.inverse_transform(y_test_cpu)
            predicted_prices = target_scaler.inverse_transform(
                test_predictions_scaled
            )

            # calculate metrics
            final_test_mse = mean_squared_error(actual_prices, predicted_prices)
            mae = mean_absolute_error(actual_prices, predicted_prices)
            r2 = r2_score(actual_prices, predicted_prices)

            print(
                f"test mse (mean squared error): ${final_test_mse:,.2f} (thousands squared)"
            )
            print(f"mean absolute error (mae): ${mae:,.2f} thousands")
            print(f"r-squared: {r2:.4f}")

    except Exception as e:
        print(f"\nan error occured during training: {str(e)}")
        print("training terminated prematurely")


def save_model_as_onnx(
    model, onnx_path, feature_scaler, target_scaler, scaler_path
):
    # set model to evaluation mode before export
    model.eval()

    # create a dummy input tensor for onnx export and move to device
    dummy_input = torch.randn(1, 3, requires_grad=True).to(DEVICE)

    try:
        torch.onnx.export(
            model,
            dummy_input,
            onnx_path,
            export_params=True,
            opset_version=18,
            do_constant_folding=True,
            # the names to assign to the input nodes of the graph
            input_names=["input"],
            # the names to assign to the output nodes to the graph
            output_names=["output"],
            dynamic_axes={
                "input": {0: "batch_size"},
                "output": {0: "batch_size"},
            },
        )

        print(f"movel successfully saved to onnx format at: {onnx_path}")

        # save the scalers
        with open(scaler_path, "wb") as f:
            pickle.dump(
                {
                    "feature_scaler": feature_scaler,
                    "target_scaler": target_scaler,
                },
                f,
            )

        print(f"scalers successfully saved to: {scaler_path}")
    except Exception as e:
        print(f"error saving model to onnx: {e}")


def predict_with_onnx(onnx_path, scaler_path, input_features_str):
    if not os.path.exists(onnx_path):
        print(f"error: onnx model file not found at {onnx_path}")
        return

    if not os.path.exists(scaler_path):
        print(f"error: scaler file not found at {scaler_path}")
        return

    # load scalers
    try:
        with open(scaler_path, "rb") as f:
            scalers = pickle.load(f)

        feature_scaler = scalers["feature_scaler"]
        target_scaler = scalers["target_scaler"]
    except Exception as e:
        print(f"error loading scalers: {e}")
        return

    # parse and validate input features
    try:
        input_values = [float(x.strip()) for x in input_features_str.split(",")]

        if len(input_values) != 3:
            raise ValueError(
                "Expected 3 features (square_footage, bedrooms, bathrooms)"
            )

        sq_footage = input_values[0]
        bedrooms = input_values[1]
        bathrooms = input_values[2]

        # basic validation
        if sq_footage <= 0:
            raise ValueError("Square footage must be a positive number")

        if not (bedrooms > 0 and int(bedrooms) == bedrooms):
            raise ValueError("Bedrooms must be a positive integer")

        if not (bathrooms > 0 and int(bathrooms) == bathrooms):
            raise ValueError("Bathrooms must be a positive integer")

        input_array = np.array(input_values).reshape(1, -1)

        # scale the input features using the loaded scaler
        scaled_input = feature_scaler.transform(input_array).astype(np.float32)

        # create an onnx runtime session
        session = ort.InferenceSession(onnx_path)
        input_name = session.get_inputs()[0].name
        output_name = session.get_outputs()[0].name

        # run inference
        try:
            ort_inputs = {input_name: scaled_input}
            ort_outputs = session.run([output_name], ort_inputs)
            predicted_scaled_price = ort_outputs[0]

            # inverse transform the predicted price to original scale
            predicted_price_thousands = target_scaler.inverse_transform(
                predicted_scaled_price
            )

            # convert to actual dollar amount
            predicted_actual_dollars = predicted_price_thousands[0][0] * 1000

            print(
                f"\nInput features: Square footage: {sq_footage}, "
                f"Bedrooms: {int(bedrooms)}, Bathrooms: {int(bathrooms)}"
            )

            # display as actual dollar amount
            print(f"Predicted price: ${predicted_actual_dollars:,.2f}")
        except Exception as e:
            print(f"error during onnx inference: {e}")

    except ValueError as e:
        print(f"error parsing or validate input features: {e}")
        print(
            "please provide features as a comma-separated list (e.g. 2500,4,2)"
        )
        return


if __name__ == "__main__":
    # command line flags
    parser = argparse.ArgumentParser(
        description="A simple neural network for housing price prediction using"
        " pytorch and onnx."
    )
    parser.add_argument(
        "--train",
        action="store_true",
        help="train the model and save it as onnx file",
    )
    parser.add_argument(
        "--predict",
        action="store_true",
        help="load an onnx model and make a prediction",
    )
    parser.add_argument(
        "--data_path",
        type=str,
        default=DATA_FILE,
        help=f"path to the csv file (default: {DATA_FILE})",
    )
    parser.add_argument(
        "--model_path",
        type=str,
        default=MODEL_ONNX_FILE,
        help=f"path to save/load the onnx model (default: {MODEL_ONNX_FILE})",
    )
    parser.add_argument(
        "--scaler_path",
        type=str,
        default=SCALER_FILE,
        help=f"path to save/load the minmax scaler (default: {SCALER_FILE})",
    )
    parser.add_argument(
        "--input_features",
        type=str,
        help="comma-separated input feature for predictions (e.g. '2500,4,2')"
        " required with --predict",
    )
    parser.add_argument(
        "--epochs",
        type=int,
        default=1000,
        help="number of training epochs (default: 1000)."
        " only applicable with --train",
    )
    parser.add_argument(
        "--lr",
        type=float,
        default=0.001,
        help="learning rate for training (default: 0.001)."
        " only applicable with --train",
    )
    parser.add_argument(
        "-patience",
        type=int,
        default=50,
        help="number of epochs to wait for improvement before early stopping"
        " (default: 50). only applicable with --train",
    )
    parser.add_argument(
        "--min_delta",
        type=float,
        default=0.0001,
        help="minimum change in test loss to qualify as an improvement for"
        " early stopping (default: 0.0001). only applicable with --train",
    )

    args = parser.parse_args()

    if args.train:
        print("--- training mode ---")

        (
            X_train,
            y_train,
            X_val,
            y_val,
            X_test,
            y_test,
            feature_scaler,
            target_scaler,
        ) = load_and_preprocess_data(args.data_path)

        if X_train is not None:
            # get a model for house price prediction
            model = HousePricePredictor().to(DEVICE)

            # train the model
            train_model(
                model,
                X_train,
                y_train,
                X_val,
                y_val,
                X_test,
                y_test,
                target_scaler,
                args.epochs,
                args.lr,
                args.patience,
                args.min_delta,
            )

            # save the model as an onnx file
            save_model_as_onnx(
                model,
                args.model_path,
                feature_scaler,
                target_scaler,
                args.scaler_path,
            )
        else:
            print("training aborted due to data loading issues")

    elif args.predict:
        print("--- predction mode ---")

        if args.input_features:
            predict_with_onnx(
                args.model_path, args.scaler_path, args.input_features
            )
        else:
            print("error: --input_features is required for prediction mode.")
            parser.print_help()
    else:
        print("please specify either --train or --predict")
        parser.print_help()
