import os
import numpy as np
import pandas as pd
from config import CONFIG
from dataclasses import dataclass
from housing_analysis.logging_utils import logger
from sklearn.model_selection import train_test_split
from housing_analysis.exceptions import DataProcessingError


@dataclass
class ModelData:
    X_train: np.ndarray
    X_test: np.ndarray
    y_train: np.ndarray
    y_test: np.ndarray


def load_data(file_path: str) -> pd.DataFrame:
    # check if file exists
    if not os.path.isfile(file_path):
        error_msg = f"File does not exist: {file_path}"
        logger.error(error_msg)
        raise DataProcessingError(error_msg)

    try:
        # load the data
        logger.info(f"Loading data from {file_path}")
        df = pd.read_csv(file_path)

        # validate data
        missing_colums = set(CONFIG["required_columns"]).difference(df.columns)
        if missing_colums:
            error_msg = f"Required columns missing: {', '.join(missing_colums)}"
            logger.error(error_msg)
            raise DataProcessingError(error_msg)
    except Exception as e:
        error_msg = f"Error loading data: {str(e)}"
        logger.error(error_msg)
        raise DataProcessingError(error_msg)

    return df


def preprocess_data(df: pd.DataFrame) -> pd.DataFrame:
    logger.info("Preprocessing data...")

    processed_df = df.copy()

    # convert required columns to numeric in one step
    for col in CONFIG["required_columns"]:
        processed_df[col] = pd.to_numeric(processed_df[col], errors="coerce")

    # handle missing values after conversion to numeric
    if processed_df[CONFIG["required_columns"]].isna().any().any():
        logger.warning("Missing values found, dropping rows")
        processed_df = processed_df.dropna(subset=CONFIG["required_columns"])

    # handle outliers
    for col in CONFIG["required_columns"]:
        # calculate mean
        mean = processed_df[col].mean()

        # get std
        std = processed_df[col].std()

        # define thresholds
        threshold = CONFIG["outlier_threshold"]
        lower_bound = mean - threshold * std
        upper_bound = mean + threshold * std

        outliers = (processed_df[col] < lower_bound) | (
            processed_df[col] > upper_bound
        )
        if outliers.any():
            logger.warning(f"Removing {outliers.sum()} outliers from {col}")
            processed_df = processed_df[~outliers]

    return processed_df


def prepare_model_data(df: pd.DataFrame) -> ModelData:
    # get X and y
    X = df[CONFIG["feature_columns"]].values
    y = df[CONFIG["target_column"]].values

    # split data into training and test sets
    X_train, X_test, y_train, y_test = train_test_split(
        X,
        y,
        test_size=CONFIG["test_size"],
        random_state=CONFIG["random_state"],
    )

    logger.info(
        f"Data split: {len(X_train)} traning samples,"
        f" {len(X_test)} test samples"
    )

    return ModelData(
        X_train=X_train,
        X_test=X_test,
        y_train=y_train,
        y_test=y_test,
    )


def make_predictions(df: pd.DataFrame, model, scaler) -> np.ndarray:
    try:
        X = df[CONFIG["feature_columns"]].values

        X_scaled = scaler.transform(X)

        predictions = model.predict(X_scaled)

        return predictions
    except Exception as e:
        error_msg = f"Error making predictions: {str(e)}"
        logger.error(error_msg)
        raise DataProcessingError(error_msg)
