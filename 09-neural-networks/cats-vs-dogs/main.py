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


def main() -> None:
    args = parse_args()

    device = setup_device()


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
        "--weigth_decay",
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
        help="Path to the image file for inference"
    )

    parser.add_argument(
        "--model_file",
        type=str,
        default=None,
        help="Path to the model file (.pth or .onnx) for inference"
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


if __name__ == "__main__":
    main()
