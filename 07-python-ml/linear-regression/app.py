import os
import sys
import logging
import argparse
import numpy as np
import pandas as pd
import matplotlib.pyplot as plt
from sklearn.preprocessing import StandardScaler
from sklearn.linear_model import LinearRegression
from sklearn.model_selection import train_test_split
from sklearn.metrics import r2_score, mean_squared_error


CONFIG = {
    "default_csv": "house_data.csv",
    "test_size": 0.2,  # 20% of data used for testing, 80% for training
    "random_state": 42,
    "figure_size": (10, 6),
    "point_color": "blue",
    "line_color": "red",
    "grid_alpha": 0.3,
    "output_image": "housing_regression.png",
    "line_width": 2,
}

logging.basicConfig(
    level=logging.INFO, format="%(asctime)s - %(levelname)s - %(message)s"
)

logger = logging.getLogger(__name__)


def main():
    # parsing command line arguments
    args = parse_arguments()

    # load and preprocess data
    df = load_data(args.file)
    processed_df = preprocess_data(df)

    # prepare data for modeling
    X = processed_df["square_footage"].to_numpy().reshape(-1, 1)  # 2d array
    y = processed_df["price_thousands"].to_numpy()

    # split data into training and test sets
    X_train, X_test, y_train, y_test = train_test_split(
        X, y, test_size=CONFIG["test_size"], random_state=CONFIG["random_state"]
    )

    logger.info(
        f"Data split: {len(X_train)} training samples, "
        f"{len(X_train)} test samples"
    )

    # train a model
    model, scaler = train_model(X_train, y_train)
    logger.info("Model training complete!")

    # evaluate model on both trainig and test data
    train_predictions, train_r2, train_rmse = evaluate_model(
        model, X_train, y_train, scaler
    )
    test_predictions, test_r2, test_rmse = evaluate_model(
        model, X_test, y_test, scaler
    )

    logger.info(
        "Model evaluation complete: "
        f"r-squared (train): {train_r2:.4f}, "
        f"r-squared (test): {test_r2:.4f}"
    )

    # print results
    print_results(
        X_train,
        y_train,
        X_test,
        y_test,
        train_predictions,
        test_predictions,
        model,
        scaler,
    )

    # create a visualization
    create_visualization(
        X_train,
        y_train,
        X_test,
        y_test,
        train_predictions,
        test_predictions,
        model,
        scaler,
        CONFIG["output_image"],
        not args.no_plot,
    )

    # predict price for houses not in our dataset
    if args.predict_sqft is not None:
        sqft_predict = args.predict_sqft
        logger.info(
            f"Predicting price for a house with {sqft_predict} square footage"
        )

        predicted_price = predict_price(model, scaler, sqft_predict)
        print(
            f"\nPredicted price for a house with {sqft_predict} square footage:"
            f" ${predicted_price:.2f} thousand"
        )
        print(
            f"This is equivalnet to approximately ${predicted_price * 1000:.2f}"
        )


def parse_arguments():
    parser = argparse.ArgumentParser(
        description="Linear Regression Analysis on Housing Data."
    )
    parser.add_argument(
        "-f",
        "--file",
        type=str,
        default=CONFIG["default_csv"],
        help=f"Path to csv file (default: {CONFIG['default_csv']})",
    )
    parser.add_argument(
        "--no-plot",
        action="store_true",
        help="Do not display the plot (still saves to file)",
    )
    parser.add_argument(
        "-predict",
        "--predict-sqft",
        type=float,
        help="Predict the price for a house with the given square footage",
    )

    return parser.parse_args()


def load_data(file_path: str) -> pd.DataFrame:
    if not os.path.isfile(file_path):
        logger.error(f"File does not exist: {file_path}")
        sys.exit(1)

    try:
        logger.info(f"Loading data from {file_path}")
        df = pd.read_csv(file_path)

        # validate for required columns
        required_columns = ["square_footage", "price_thousands"]
        for col in required_columns:
            if col not in df.columns:
                logger.error(f"Required column {col} not found in the csv file")
                sys.exit(1)

        return df
    except Exception as e:
        logger.error(f"Error loading data: {str(e)}")
        sys.exit(1)


def preprocess_data(df: pd.DataFrame) -> pd.DataFrame:
    logger.info("Preprocessing data...")
    processed_df = df.copy()

    # handle missing data
    if processed_df[["square_footage", "price_thousands"]].isna().any().any():
        logger.warning(
            "missing values found, dropping rows with missing values"
        )
        processed_df = processed_df.dropna(
            subset=["square_footage", "price_thousands"]
        )

    # filter out outliers
    for col in ["square_footage", "price_thousands"]:
        mean = processed_df[col].mean()
        std = processed_df[col].std()
        lower_bound = mean - 3 * std
        upper_bound = mean + 3 * std

        outliers = (processed_df[col] < lower_bound) | (
            processed_df[col] > upper_bound
        )
        if outliers.any():
            logger.warning(f"Removing {outliers.sum()} outliers from {col}")
            processed_df = processed_df[~outliers]

    # ensure numeric types
    processed_df["square_footage"] = pd.to_numeric(
        processed_df["square_footage"], errors="coerce"
    )
    processed_df["price_thousands"] = pd.to_numeric(
        processed_df["price_thousands"], errors="coerce"
    )

    processed_df = processed_df.dropna(
        subset=["square_footage", "price_thousands"]
    )

    return processed_df


def train_model(X, y):
    # scale features
    scaler = StandardScaler()
    X_scaled = scaler.fit_transform(X)

    # train the model
    model = LinearRegression()
    model.fit(X_scaled, y)

    return model, scaler


def evaluate_model(model: LinearRegression, X, y, scaler: StandardScaler):
    X_scaled = scaler.transform(X)
    predictions = model.predict(X_scaled)

    # calculate quality metrics
    r_squared = r2_score(y, predictions)
    rmse = np.sqrt(mean_squared_error(y, predictions))

    return predictions, r_squared, rmse


def print_results(
    X_train,
    y_train,
    X_test,
    y_test,
    train_predictions,
    test_predictions,
    model: LinearRegression,
    scaler: StandardScaler,
):
    assert scaler.scale_ is not None
    assert scaler.mean_ is not None

    slope = model.coef_[0] / scaler.scale_[0]
    intercept = model.intercept_ - (
        model.coef_[0] * scaler.mean_[0] / scaler.scale_[0]
    )

    r_squared_train = r2_score(y_train, train_predictions)
    r_squared_test = r2_score(y_test, test_predictions)
    rmse_train = np.sqrt(mean_squared_error(y_train, train_predictions))
    rmse_test = np.sqrt(mean_squared_error(y_test, test_predictions))

    print(
        "\nLinear Regression Formula: "
        f"Price = {slope:.4f} x Square Footage + {intercept:.4f}"
    )
    print(f"R-squared (training): {r_squared_train:.4f}")
    print(f"R-squared (test): {r_squared_test:.4f}")
    print(f"RMSE (training): {rmse_train:.4f}")
    print(f"RMSE (test): {rmse_test:.4f}")

    train_df = pd.DataFrame(
        {
            "Square Footage": X_train.flatten(),
            "Actual Price ($K)": y_train,
            "Predicted Price ($K)": np.round(train_predictions, 2),
        }
    )

    test_df = pd.DataFrame(
        {
            "Square Footage": X_test.flatten(),
            "Actual Price ($K)": y_test,
            "Predicted Price ($K)": np.round(test_predictions, 2),
        }
    )

    print("\nTraining Prediction Sample (first 5 rows):")
    print(train_df.head().to_string(index=False))
    print("\nTest Prediction Sample (first 5 rows):")
    print(test_df.head().to_string(index=False))


def create_visualization(
    X_train,
    y_train,
    X_test,
    y_test,
    train_predictions,
    test_predictions,
    model: LinearRegression,
    scaler: StandardScaler,
    output_file,
    show_plot=True,
):
    plt.figure(figsize=CONFIG["figure_size"])

    # plot the training data
    plt.scatter(
        X_train,
        y_train,
        color=CONFIG["point_color"],
        alpha=0.7,
        label="Training data",
    )

    # plot the testing data
    plt.scatter(X_test, y_test, color="green", alpha=0.7, label="Test data")

    x_range = np.linspace(
        min(X_train.min(), X_test.min()),
        max(X_train.max(), X_test.max()),
        100,
    ).reshape(-1, 1)

    # scale the range and predict corresponding y-values
    x_range_scaled = scaler.transform(x_range)
    y_range_pred = model.predict(x_range_scaled)

    # plot the regression line
    plt.plot(
        x_range,
        y_range_pred,
        color=CONFIG["line_color"],
        linewidth=CONFIG["line_width"],
        label="Regression line",
    )

    # add labels and title
    plt.xlabel("Square Footage")
    plt.ylabel("Price (thousands $)")
    plt.title("Linear Regression: Housing Price vs Square Footage")
    plt.legend()
    plt.grid(True, alpha=CONFIG["grid_alpha"])

    # calculate and display model parameters
    assert scaler.scale_ is not None
    assert scaler.mean_ is not None

    slope = model.coef_[0] / scaler.scale_[0]
    intercept = model.intercept_ - (
        model.coef_[0] * scaler.mean_[0] / scaler.scale_[0]
    )
    r_squared_train = r2_score(y_train, train_predictions)
    r_squared_test = r2_score(y_test, test_predictions)

    # format text to display on plot
    formula_text = f"Price = {slope:.4f} x Square Footage + {intercept:.4f}"
    r2_train_text = f"R2 (train): {r_squared_train:.4f}"
    r2_test_text = f"R2 (test): {r_squared_test:.4f}"

    plt.figtext(0.52, 0.18, formula_text, fontsize=12)
    plt.figtext(0.52, 0.15, r2_train_text, fontsize=12)
    plt.figtext(0.52, 0.12, r2_test_text, fontsize=12)

    plt.savefig(output_file)
    logger.info(f"Plot save as {output_file}")

    if show_plot:
        plt.show()

    plt.close()


def predict_price(model: LinearRegression, scaler: StandardScaler, sqft: float):
    sqft_arr = np.array([[sqft]])
    sqft_scaled = scaler.transform(sqft_arr)
    predicted_price = model.predict(sqft_scaled)

    return predicted_price[0]


if __name__ == "__main__":
    main()
