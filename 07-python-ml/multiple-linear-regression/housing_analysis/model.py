import os
import json
import pickle
import numpy as np
from config import CONFIG
from dataclasses import dataclass
from typing import List, Tuple, Any
from sklearn.preprocessing import StandardScaler
from sklearn.linear_model import LinearRegression
from sklearn.metrics import mean_squared_error, r2_score
from housing_analysis.data_processing import ModelData
from housing_analysis.exceptions import ModelOperationError
from housing_analysis.logging_utils import logger


@dataclass
class ModelResults:
    model: LinearRegression
    scaler: StandardScaler
    train_predictions: np.ndarray
    test_predictions: np.ndarray
    train_r2: float
    test_r2: float
    train_rmse: float
    test_rmse: float


def train_model(data: ModelData) -> Tuple[LinearRegression, StandardScaler]:
    scaler = StandardScaler()
    X_scaled = scaler.fit_transform(data.X_train)

    # train the model
    model = LinearRegression()
    model.fit(X_scaled, data.y_train)

    return model, scaler


def evaluate_model(
    data: ModelData, model: LinearRegression, scaler: StandardScaler
) -> ModelResults:
    # evaluate our training data
    X_train_scaled = scaler.transform(data.X_train)
    train_predictions = model.predict(X_train_scaled)

    train_r2 = r2_score(data.y_train, train_predictions)
    train_rmse = np.sqrt(mean_squared_error(data.y_train, train_predictions))

    # evaluate our test data
    X_test_scaled = scaler.transform(data.X_test)
    test_predictions = model.predict(X_test_scaled)

    test_r2 = r2_score(data.y_test, test_predictions)
    test_rmse = np.sqrt(mean_squared_error(data.y_test, test_predictions))

    logger.info(
        f"Model evaluate: R2 (train): {train_r2:.4f}, R2 (test): {test_r2:.4f}"
    )

    return ModelResults(
        model=model,
        scaler=scaler,
        train_predictions=train_predictions,
        test_predictions=test_predictions,
        train_r2=train_r2,
        test_r2=test_r2,
        train_rmse=train_rmse,
        test_rmse=test_rmse,
    )


def save_model(
    model_results: ModelResults, model_path: str, metadata_path: str
) -> None:
    try:
        # create directory if it not exists
        model_dir = os.path.dirname(model_path)
        if model_dir and not os.path.exists(model_dir):
            os.makedirs(model_dir)

        metadata_dir = os.path.dirname(metadata_path)
        if metadata_dir and not os.path.exists(metadata_dir):
            os.makedirs(metadata_dir)

        # save model and scaler
        with open(model_path, "wb") as f:
            model_components = {
                "model": model_results.model,
                "scaler": model_results.scaler,
            }
            pickle.dump(model_components, f)

        intercept, coefficients = get_model_formula(model_results)
        metadata = {
            "intercept": intercept,
            "coefficients": coefficients,
            "feature_names": CONFIG["feature_columns"],
            "target_name": CONFIG["target_column"],
            "train_r2": float(model_results.train_r2),
            "test_r2": float(model_results.test_r2),
            "train_rmse": float(model_results.train_rmse),
            "test_rmse": float(model_results.test_rmse),
        }

        with open(metadata_path, "w") as f:
            json.dump(metadata, f, indent=4)

        logger.info(f"Model saved to {model_path}")
        logger.info(f"Model metadata saved to {metadata_path}")

    except Exception as e:
        error_msg = f"Error saving model: {str(e)}"
        logger.error(error_msg)
        raise ModelOperationError(error_msg)


def get_model_formula(model_results: ModelResults) -> Tuple[float, List[float]]:
    model = model_results.model
    scaler = model_results.scaler

    coefficients = []
    scale = np.array(scaler.scale_)
    mean = np.array(scaler.mean_)

    for i in range(len(model.coef_)):
        coef = model.coef_[i] / scale[i]
        coefficients.append(float(coef))

    intercept = model.intercept_ - sum(
        model.coef_[i] * mean[i] / scale[i] for i in range(len(model.coef_))
    )

    return float(intercept), coefficients


def load_model(
    model_path: str, metadata_path: str = ""
) -> Tuple[LinearRegression, StandardScaler, dict[str, Any]]:
    try:
        if not os.path.isfile(model_path):
            error_msg = f"Model file does not exist: {model_path}"
            logger.error(error_msg)
            raise ModelOperationError(error_msg)

        with open(model_path, "rb") as f:
            model_components = pickle.load(f)

        model = model_components["model"]
        scaler = model_components["scaler"]

        metadata = {}
        if metadata_path and os.path.isfile(metadata_path):
            with open(metadata_path, "r") as f:
                metadata = json.load(f)

        logger.info(f'Model loaded from {model_path}')
        if metadata:
            logger.info(f'Model metadata loaded from {metadata_path}')

        return model, scaler, metadata
    except Exception as e:
        error_msg = f'Error loading model: {str(e)}'
        logger.error(error_msg)
        raise ModelOperationError(error_msg)