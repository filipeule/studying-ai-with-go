import gc
import os
import argparse
import warnings
from pathlib import Path

import torch
import torch.nn as nn
import torch.optim as optim
from torch.utils.data import DataLoader, random_split, Subset
from torchvision import datasets, transforms
from tqdm import tqdm


class CatDogCNN(nn.Module):
    def __init__(self, image_size) -> None:
        super(CatDogCNN, self).__init__()

        # first convolutional block
        # conv2d: applies 2d convolution to extract visual features
        # 3 input channels (RGB), 32 output feature maps, 3x3 kernel
        self.conv1 = nn.Conv2d(3, 32, kernel_size=3, padding=1)

        # batchnorm2d: normalizes the output of the convolutional layer
        # helps with faster and more stable training
        self.bn1 = nn.BatchNorm2d(32)

        # maxpool2d: reduces spatial dimensions by taking maximum in 2x2 regions
        # this reduce computation and helps with translation invariance
        self.pool1 = nn.MaxPool2d(kernel_size=2, stride=2)

        # second convolutional block
        self.conv2 = nn.Conv2d(32, 64, kernel_size=3, padding=1)
        self.bn2 = nn.BatchNorm2d(64)
        self.pool2 = nn.MaxPool2d(kernel_size=2, stride=2)

        # third convolutional block
        self.conv3 = nn.Conv2d(64, 128, kernel_size=3, padding=1)
        self.bn3 = nn.BatchNorm2d(128)
        self.pool3 = nn.MaxPool2d(kernel_size=2, stride=2)

        # forth convolutional block
        self.conv4 = nn.Conv2d(128, 256, kernel_size=3, padding=1)
        self.bn4 = nn.BatchNorm2d(256)
        self.pool4 = nn.MaxPool2d(kernel_size=2, stride=2)

        # calculate dimensions for a fully connected layer
        # after 4 pooling layers (each reducing dimensions by half)
        # the size is divided by 16
        feature_size = image_size // 16
        fc_input_size = 256 * feature_size * feature_size

        # fully connected layers for classification
        # takes flatten features maps and outputs class probabilities
        self.fc1 = nn.Linear(fc_input_size, 512)
        # dropout layer: randomly zeroes some elements during training
        # this prevent overfitting by making the network more robust
        self.dropout = nn.Dropout(0.5)

        # final layer outputs 2 values: one per class
        self.fc2 = nn.Linear(512, 2)
        self.relu = nn.ReLU()

    def forward(self, x):
        # convolutional feature extration
        for i in range(1, 5):
            # dynamically get layers for the current block (conv1, bn1, pool1, etc)
            # getattr(self, f'conv{i}')
            conv = getattr(self, f"conv{i}")
            bn = getattr(self, f"bn{i}")
            pool = getattr(self, f"pool{i}")

            # step 1: apply 2d convolution
            x = conv(x)

            # step 2: apply batch normalization
            x = bn(x)

            # step 3: apply relu activation
            x = self.relu(x)

            # step 4: apply max pooling
            x = pool(x)

        # after all blocks, tensor shape is approx. [batch, 256, image_size/16]
        # prepare for classification
        # flatten 4d feature maps into 2d for fully connected layers
        x = x.view(x.size(0), -1)

        # classification layers
        # first fully connected layer
        x = self.fc1(x)

        # apply relu activation
        x = self.relu(x)

        # dropout layer
        x = self.dropout(x)

        # final classification layer
        x = self.fc2(x)

        return x


class EarlyStopping:
    def __init__(self, patience=3, min_delta=0.001) -> None:
        self.patience = patience
        self.min_delta = min_delta
        self.counter = 0
        self.best_score = None
        self.early_stop = False

    def __call__(self, val_loss):
        score = -val_loss

        if self.best_score is None:
            self.best_score = score
        elif score < self.best_score + self.min_delta:
            self.counter += 1
            print(
                f"Early stopping counter: {self.counter} out of {self.patience}"
            )
            if self.counter >= self.patience:
                self.early_stop = True
                print("Early stopping triggered!")
        else:
            self.best_score = score
            self.counter = 0

        return self.early_stop


def main() -> None:
    setup_warning_supression()

    args = parse_args()
    device = setup_device()

    if args.inference:
        print("Performing inference")

        if not args.image_path:
            print("Error: --image_path is required for inference")
            exit(1)

        if args.model_file:
            model_file = args.model_file
        elif Path(args.model_path).exists():
            model_file = args.model_path
            print(f"Using default PyTorch model: {model_file}")
        elif Path(args.onnx_path).exists():
            model_file = args.onnx_path
            print(f"Using default onnx model: {model_file}")
        else:
            print("Error: no trained model found")
            exit(1)

        run_inference(args.image_path, model_file, args.image_size, device)
        return

    else:
        print("Training model")

        using_workers = run_training_and_cleanup(args, device)

        if using_workers:
            print("Forcing clean exit...")
            os._exit(0)


def setup_warning_supression():
    import warnings
    import PIL.Image

    warnings.filterwarnings("ignore", message="Truncated File Read")
    warnings.filterwarnings(
        "ignore", message=".*Truncated.*", category=UserWarning
    )
    warnings.filterwarnings(
        "ignore", category=UserWarning, module="PIL.TiffImagePlugin"
    )
    warnings.filterwarnings(
        "ignore", message=".*EXIF.*", category=UserWarning
    )
    warnings.filterwarnings(
        "ignore", message=".*palette.*", category=UserWarning
    )
    PIL.Image.warnings.simplefilter(
        "ignore", PIL.Image.DecompressionBombWarning
    )


def parse_args() -> argparse.Namespace:
    parser = argparse.ArgumentParser(
        description="Train a cat vs dog classifier"
    )

    parser.add_argument(
        "--data_dir",
        type=str,
        default="./data",
        help="Path to the dataset directory",
    )

    parser.add_argument(
        "--image_size",
        type=int,
        default=256,
        help="Size to resize images (square) - larger values prove mode detail"
        " , but take more memory",
    )

    parser.add_argument(
        "--augmentation",
        action="store_true",
        help="Enable data augmentation for training (applies transforms to"
        " artificially increase data set size)",
    )

    parser.add_argument(
        "--batch_size",
        type=int,
        default=32,
        help="Batch size for traininig - how many images to process at once",
    )

    parser.add_argument(
        "--learning_rate",
        type=float,
        default=0.001,
        help="Learning rate - controls how quickly our model learns",
    )

    parser.add_argument(
        "--num_epochs",
        type=int,
        default=10,
        help="Number of training epochs - each epoch processes the entire"
        " dataset once",
    )

    parser.add_argument(
        "--momentum",
        type=float,
        default=0.9,
        help="Momentum for SGD optimizer - helps accelerate in relevant"
        " directions and dampen oscillations",
    )

    parser.add_argument(
        "--weight_decay",
        type=float,
        default=1e-4,
        help="Weight decay (L2 penalty) - helps prevents overfeeding by"
        " penalizing large weights",
    )

    parser.add_argument(
        "--model_path",
        type=str,
        default="cat_dog_classifier.pth",
        help="Path to save the PyTorch model",
    )

    parser.add_argument(
        "--onnx_path",
        type=str,
        default="cat_dog_classifier.onnx",
        help="Path to save the onnx model (a format for model interoperability)",
    )

    parser.add_argument(
        "--val_split",
        type=float,
        default=0.2,
        help="Validation set split ratio - percentage of data used for"
        " validation",
    )

    parser.add_argument(
        "--patience",
        type=int,
        default=2,
        help="Patience for learning rate scheduler - how many epochs to wait"
        " before reducing learning rate",
    )

    parser.add_argument(
        "--early_stopping", action="store_true", help="Enable early stopping"
    )

    parser.add_argument(
        "--early_stopping_patience",
        type=int,
        default=3,
        help="Number of epochs to wait before stop training if no improvement",
    )

    parser.add_argument(
        "--early_stopping_min_delta",
        type=float,
        default=0.001,
        help="Minimum change to qualify as improvement",
    )

    parser.add_argument(
        "--inference",
        action="store_true",
        help="Run inference on a single image instead of training",
    )

    parser.add_argument(
        "--image_path",
        type=str,
        default=None,
        help="Path to the image file for inference",
    )

    parser.add_argument(
        "--model_file",
        type=str,
        default=None,
        help="Path to the model file (.pth or .onnx) for inference",
    )

    return parser.parse_args()


def setup_device() -> torch.device:
    if torch.cuda.is_available():
        device = torch.device("cuda")
        print("Using NVIDIA GPU (CUDA)")
    elif hasattr(torch.backends, "mps") and torch.backends.mps.is_available():
        device = torch.device("mps")
        print("Using Apple GPU (MPS)")
    else:
        device = torch.device("cpu")
        print("Using CPU")

    return device


def run_inference(image_path, model_file, image_size, device):
    import numpy as np
    from PIL import Image

    if not Path(image_path).exists() or not Path(model_file).exists():
        print("Error: image or model not found")
        return None, None

    transform = transforms.Compose(
        [
            transforms.Resize((image_size, image_size)),
            transforms.ToTensor(),
            transforms.Normalize(
                mean=[0.485, 0.456, 0.406], std=[0.229, 0.224, 0.225]
            ),
        ]
    )

    try:
        image = Image.open(image_path).convert("RGB")
        input_tensor = transform(image).unsqueeze(0).to(device)
    except Exception as e:
        print(f"Error loading image: {e}")
        return None, None

    if Path(model_file).suffix == ".pth":
        model = CatDogCNN(image_size).to(device)

        try:
            model.load_state_dict(torch.load(model_file, map_location=device))
            model.eval()

            with torch.no_grad():
                outputs = model(input_tensor)
                probabilities = torch.nn.functional.softmax(outputs, dim=1)
                confidence, predicted = torch.max(probabilities, 1)
                pred_class = ["cat", "dog"][predicted.item()]
                conf_score = confidence.item()
        except Exception as e:
            print(f"Error with PyTorch model: {e}")
            return None, None
    elif Path(model_file).suffix == ".onnx":
        try:
            import onnxruntime as ort

            ort_session = ort.InferenceSession(model_file)
            ort_inputs = {
                ort_session.get_inputs()[0].name: input_tensor.cpu().numpy()
            }
            ort_outputs = ort_session.run(None, ort_inputs)
            outputs = ort_outputs[0]

            def softmax(x):
                exp_x = np.exp(x - np.max(x))
                return exp_x / exp_x.sum()

            probabilities = softmax(outputs[0])
            predicted = np.argmax(probabilities)
            pred_class = ["cat", "dog"][predicted]
            conf_score = float(probabilities[predicted])

        except Exception as e:
            print(f"Error with ONNX model: {e}")
            return None, None
    else:
        print("Error: unsupported model format")
        return None, None

    print(
        f"\nInference results:\nImage: {image_path}\nPrediction: {pred_class}\nConfidence: {conf_score:.2f}"
    )

    return pred_class, conf_score


def run_training_and_cleanup(args, device):
    using_workers = False

    try:
        train_loader, val_loader = load_data(
            args.data_dir,
            args.image_size,
            args.batch_size,
            args.val_split,
            device,
            args.augmentation,
        )

        print("Data loading completed successfully!")
        using_workers = True

        print("Initializing neural network model...")
        model = CatDogCNN(args.image_size).to(device)

        print("Model architecture:")
        print(model)

        # calculate and display total number of trainable parameters
        # this gives us insight into model complexity and memory requirements
        total_params = sum(
            p.numel() for p in model.parameters() if p.requires_grad
        )
        print(f"Total trainable parameters: {total_params:,}")

        # configure loss function
        print("Setting up training components...")
        criterion = nn.CrossEntropyLoss()

        # configure is the optimizer
        optimizer = optim.SGD(
            model.parameters(),  # all model weights and biases to optimize
            lr=args.learning_rate,  # step size for weight updates
            momentum=args.momentum,  # momentum factor for smoother convergence
            weight_decay=args.weight_decay,  # regularization
        )

        # learning rate scheduler
        scheduler = optim.lr_scheduler.ReduceLROnPlateau(
            optimizer,
            mode="min",
            factor=0.1,
            patience=args.patience,
        )

        # training loop execution
        print("Starting training process...")
        print(f"Training for maximum {args.num_epochs} epochs...")
        if args.early_stopping:
            print(
                "Early stopping enable: will stop if no improvement for"
                f" {args.early_stopping_patience} epochs"
            )

        best_model_state, best_val_accuracy = train_model(
            model,
            train_loader,
            val_loader,
            criterion,
            optimizer,
            scheduler,
            device,
            args.num_epochs,
            args.early_stopping,
            args.early_stopping_patience,
            args.early_stopping_min_delta,
        )

        # best model restoration
        if best_model_state is not None:
            model.load_state_dict(best_model_state)
            print(
                f"Restored best model state (validation accuracy: {best_val_accuracy:.2f})%"
            )
        else:
            print("Warning: No best model state saved, using final epoch model")

        # model persistence
        print("Saving training model...")
        save_model(
            model, args.model_path, args.onnx_path, args.image_size, device
        )
        print("Model saved sucessfully!")
        print("\n" + "=" * 50)
        print("ALL TRAINING OPERATIONS COMPLETE SUCCESSFULLY!")
        print("=" * 50)

        # immediate cleanup of large objects for smooth operation
        del train_loader
        del val_loader
        del model

        return using_workers

    except Exception as e:
        print(f"\nERROR DURING TRAINING: {e}")
        print("Proceeding with cleanup and resource deallocation")
        return using_workers
    finally:
        print("Cleaning resources...")
        if device.type == "cuda":
            torch.cuda.empty_cache()
        elif device.type == "mps":
            gc.collect()
            gc.collect()
        gc.collect()


def load_data(
    data_dir, image_size, batch_size, val_split, device, augmentation
):
    # warnings.filterwarnings(
    #     "ignore",
    #     message="Truncated File Read",
    #     category=UserWarning,
    #     module="PIL.TiffImagePlugin",
    # )
    warnings.filterwarnings(
        "ignore", category=UserWarning, module="PIL.TiffImagePlugin"
    )

    use_pin_memory = device.type != "mps"

    val_transform = transforms.Compose(
        [
            transforms.Resize((image_size, image_size)),
            transforms.ToTensor(),
            transforms.Normalize(
                mean=[0.485, 0.456, 0.406], std=[0.229, 0.224, 0.225]
            ),
        ]
    )

    if augmentation:
        print("using data augmentation during training")

        train_transform = transforms.Compose(
            [
                transforms.Resize((image_size, image_size)),
                transforms.RandomHorizontalFlip(p=0.5),
                transforms.RandomRotation(15),
                transforms.ColorJitter(
                    brightness=0.2, contrast=0.2, saturation=0.2, hue=0.1
                ),
                transforms.RandomAffine(
                    degrees=0, translate=(0.1, 0.1), scale=(0.9, 1.1)
                ),
                transforms.ToTensor(),
                transforms.Normalize(
                    mean=[0.485, 0.456, 0.406], std=[0.229, 0.224, 0.225]
                ),
            ]
        )
    else:
        print("No data augmentation")

        train_transform = val_transform

    try:
        full_dataset = datasets.ImageFolder(root=data_dir)

        print(f"Found {len(full_dataset)} images in total")
        print(f"Classes: {full_dataset.classes}")

        train_size = int((1 - val_split) * len(full_dataset))
        val_size = len(full_dataset) - train_size

        generator = torch.Generator().manual_seed(42)

        train_indices, val_indices = random_split(
            range(len(full_dataset)),
            [train_size, val_size],
            generator=generator,
        )

        train_subset = Subset(full_dataset, train_indices.indices)
        train_subset.dataset.transform = train_transform

        val_subset = Subset(full_dataset, val_indices.indices)
        val_subset.dataset.transform = val_transform

        print(f"Training set: {len(train_subset)} images")
        print(f"Validation set: {len(val_subset)} images")

    except Exception as e:
        print(f"Error loading dataset from {data_dir}: {e}")
        exit(1)

    train_loader = DataLoader(
        train_subset,
        batch_size=batch_size,
        shuffle=True,
        num_workers=4,
        pin_memory=use_pin_memory,
    )

    val_loader = DataLoader(
        val_subset,
        batch_size=batch_size,
        shuffle=False,
        num_workers=4,
        pin_memory=use_pin_memory,
    )

    return train_loader, val_loader


def train_model(
    model,
    train_loader,
    val_loader,
    criterion,
    optimizer,
    scheduler,
    device,
    num_epochs,
    early_stopping=False,
    early_stopping_patience=3,
    early_stopping_min_delta=0.001,
):
    print("\nStarting training...")

    train_loss = []
    val_accuracies = []

    best_val_accuracy = 0.0
    best_model_state = None

    early_stopper = None

    if early_stopping:
        early_stopper = EarlyStopping(
            patience=early_stopping_patience,
            min_delta=early_stopping_min_delta,
        )
        print(f"Early stopping enable with patience={early_stopping_patience}")

    for epoch in range(num_epochs):
        model.train()
        running_loss = 0.0
        train_loader_tqdm = tqdm(
            train_loader, desc=f"Epoch {epoch + 1}/{num_epochs} (Training)"
        )

        for inputs, labels in train_loader_tqdm:
            inputs, labels = inputs.to(device), labels.to(device)

            optimizer.zero_grad()
            outputs = model(inputs)
            loss = criterion(outputs, labels)
            loss.backward()
            optimizer.step()

            running_loss += loss.item()
            train_loader_tqdm.set_postfix(
                loss=running_loss / (train_loader_tqdm.n + 1)
            )

        avg_train_loss = running_loss / len(train_loader)
        train_loss.append(avg_train_loss)

        # validation phase
        model.eval()

        correct_predictions = 0
        total_samples = 0
        val_running_loss = 0.0

        with torch.no_grad():
            val_loader_tqdm = tqdm(
                val_loader, desc=f"Epoch {epoch + 1}/{num_epochs} (Validation)"
            )

            for inputs, labels in val_loader_tqdm:
                inputs, labels = inputs.to(device), labels.to(device)

                outputs = model(inputs)
                loss = criterion(outputs, labels)
                val_running_loss += loss.item()

                _, predicted = torch.max(outputs.data, 1)
                total_samples += labels.size(0)
                correct_predictions += (predicted == labels).sum().item()

                val_loader_tqdm.set_postfix(
                    accuracy=f"{100 * correct_predictions / total_samples:.2f}%"
                )

        avg_val_loss = val_running_loss / len(val_loader)
        val_accuracy = 100 * correct_predictions / total_samples
        val_accuracies.append(val_accuracy)

        scheduler.step(avg_val_loss)
        current_lr = optimizer.param_groups[0]["lr"]
        print(f"Current learning rate: {current_lr}")

        if val_accuracy > best_val_accuracy:
            best_val_accuracy = val_accuracy
            best_model_state = model.state_dict().copy()

            print(f"New best model: {best_val_accuracy:.2f}% accuracy")

        print(
            f"Epoch {epoch + 1}/{num_epochs}: Train loss={avg_train_loss:.2f},"
            f" Val loss={avg_val_loss:.4f}, Val acc={val_accuracy:.2f}%"
        )

        if early_stopping and early_stopper(avg_val_loss):
            print(f"Early stopping triggered after {epoch + 1} epochs")
            break

    print(
        f"Training finished! Best validation accuracy: {best_val_accuracy:.2f}%"
    )

    return best_model_state, best_val_accuracy


def save_model(model, model_path, onnx_path, image_size, device):
    torch.save(model.state_dict(), model_path)
    print(f"PyTorch model saved to {model_path}")

    model.eval()
    dummy_input = torch.randn(1, 3, image_size, image_size).to(device)

    try:
        with torch.no_grad():
            torch.onnx.export(
                model,
                dummy_input,
                onnx_path,
                export_params=True,
                opset_version=18,
                do_constant_folding=True,
                input_names=["input"],
                output_names=["output"],
                dynamic_axes={
                    "input": {0: "batch_size"},
                    "output": {0: "batch_size"},
                },
            )

        print(f"Model exported to onnx format at {onnx_path}")
    except Exception as e:
        print(f"Error during onnx export: {e}")


if __name__ == "__main__":
    main()
