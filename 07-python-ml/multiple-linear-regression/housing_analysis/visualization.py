import warnings
import pandas as pd
import numpy as np
import matplotlib.pyplot as plt
from typing import Dict, Any

from config import CONFIG
from housing_analysis import logger
from housing_analysis import ModelData
from housing_analysis import ModelResults, get_model_formula


def print_results(data: ModelData, model_results: ModelResults) -> None:
    # get model formula
    intercept, coefficients = get_model_formula(model_results)
    print("\nMultiple Linear Regression Formula")
    print(
        f"Price = {intercept:.4f} + {coefficients[0]:.4f} x Square Footage + {coefficients[1]:.4f} x Bedrooms"
    )

    # r2 values (training and testing)
    print(f"R-squared (training): {model_results.train_r2:.4f}")
    print(f"R-squared (test): {model_results.test_r2:.4f}")

    # rmse (training and testing)
    print(f"RMSE (training): {model_results.train_rmse:.4f}")
    print(f"RMSE (test): {model_results.test_rmse:.4f}")

    # create dataframes
    train_df = pd.DataFrame(
        {
            "Square Footage": data.X_train[:, 0],
            "Bedrooms": data.X_train[:, 1],
            "Actual Price ($K)": data.y_train,
            "Predicted Price ($K)": np.round(
                model_results.train_predictions, 2
            ),
        }
    )

    test_df = pd.DataFrame(
        {
            "Square Footage": data.X_test[:, 0],
            "Bedrooms": data.X_test[:, 1],
            "Actual Price ($K)": data.y_test,
            "Predicted Price ($K)": np.round(model_results.test_predictions, 2),
        }
    )

    # print sample of results
    print("\nTraining Data Sample (first five rows):")
    print(train_df.head().to_string(index=False))

    print("\nTest Data Sample (first five rows):")
    print(test_df.head().to_string(index=False))


def create_visualization_data(
    data: ModelData, model_results: ModelResults
) -> Dict[str, Any]:
    # combine training and test data to get the full range of values
    X_combined = np.vstack((data.X_train, data.X_test))

    # get the feature range for plotting
    x_min, x_max = (
        X_combined[:, 0].min(),
        X_combined[:, 0].max(),
    )  # square footage
    y_min, y_max = X_combined[:, 1].min(), X_combined[:, 1].max()  # bedrooms

    feature_ranges = [
        np.linspace(x_min, x_max, 1000),
        np.linspace(y_min, y_max, 1000),
    ]

    # calculate the mean of features for regression line/plane
    feature_means = X_combined.mean(axis=0)

    # get formula for display
    intercept, coefficients = get_model_formula(model_results)
    formula_text = f"Price = {intercept:.4f} + {coefficients[0]:.4f} x Square Footage + {coefficients[1]:.4f} x Bedrooms"

    # create mesh grid for 3d visualization
    x_range = np.linspace(x_min, x_max, CONFIG["mesh_grid_size"])
    y_range = np.linspace(y_min, y_max, CONFIG["mesh_grid_size"])
    xx, yy = np.meshgrid(x_range, y_range)

    # prepare grid points for prediction
    grid_points = np.c_[xx.ravel(), yy.ravel()]

    # scale grid points the same scaler used for training
    grid_points_scaled = model_results.scaler.transform(grid_points)

    # make predictions
    z_pred = model_results.model.predict(grid_points_scaled)
    zz = z_pred.reshape(xx.shape)

    return {
        "feature_ranges": feature_ranges,
        "feature_means": feature_means,
        "formula_text": formula_text,
        "xx": xx,
        "yy": yy,
        "zz": zz,
    }


def create_2d_visualization(
    data: ModelData,
    model_results: ModelResults,
    viz_data: Dict[str, Any],
    output_file: str,
    show_plot: bool = True,
) -> None:
    # create a figure with two side-by-side plot (1 row, 2 columns)
    fig, axes = plt.subplots(1, 2, figsize=CONFIG["figure_size"])

    features_names = CONFIG["feature_columns"]
    feature_ranges = viz_data["feature_ranges"]
    feature_means = viz_data["feature_means"]

    # create a plot for each feature
    for i, feature in enumerate(features_names):
        ax = axes[i]

        # extract feature values for this specific feature
        X_train_feature = data.X_train[:, i]
        X_test_feature = data.X_test[:, i]

        # plot training data points
        ax.scatter(
            X_train_feature,  # x coords
            data.y_train,  # y coords
            color=CONFIG["point_color"],
            alpha=CONFIG["point_alpha"],
            label="Training data",
        )

        ax.scatter(
            X_test_feature,  # x coords
            data.y_test,  # y coords
            color=CONFIG["test_point_color"],
            alpha=CONFIG["point_alpha"],
            label="Test data",
        )

        # add regression line
        if i == 0:  # square footage plot
            line_X = np.c_[
                feature_ranges[0],
                np.full(feature_ranges[0].shape, feature_means[1]),
            ]
        else:
            line_X = np.c_[
                np.full(feature_ranges[1].shape, feature_means[0]),
                feature_ranges[1],
            ]

        # scale the line points and predict prices
        line_X_scaled = model_results.scaler.transform(line_X)
        line_y = model_results.model.predict(line_X_scaled)

        # plot the regression line
        ax.plot(
            feature_ranges[i],
            line_y,
            color=CONFIG["line_color"],
            linewidth=CONFIG["line_width"],
            label="Regression line",
        )

        # add labels and title
        ax.set_xlabel(feature.replace("_", " ").title())
        ax.set_ylabel("Price (thousands $)")
        ax.set_title(
            f"Price vs {feature.replace('_', ' ').title()} with Regression line"
        )
        ax.legend()
        ax.grid(True, alpha=CONFIG["grid_alpha"])

    # add overall title
    plt.suptitle("Multiple Linear Regression: Housing Price vs Features")
    plt.tight_layout(rect=(0, 0, 1, 0.95))

    plt.savefig(output_file)
    logger.info(f"2d plot saved as {output_file}")

    if show_plot:
        plt.show()

    plt.close()


def create_3d_visualization(
    data: ModelData,
    model_results: ModelResults,
    viz_data: Dict[str, Any],
    output_file: str,
    show_plot: bool = True,
) -> None:
    with warnings.catch_warnings():
        warnings.filterwarnings("ignore", category=RuntimeWarning)

        # create a figure and add a 3d subplot
        fig = plt.figure(figsize=CONFIG["figure_size"])
        ax: Any = fig.add_subplot(111, projection="3d")

        # set initial view angle
        ax.view_init(elev=30, azim=45)

        # plot training data points
        ax.scatter(
            data.X_train[:, 0],
            data.X_train[:, 1],
            data.y_train,
            color=CONFIG["point_color"],
            alpha=CONFIG["point_alpha"],
            label="Training data",
        )

        # plot test data points
        ax.scatter(
            data.X_test[:, 0],
            data.X_test[:, 1],
            data.y_test,
            color=CONFIG["test_point_color"],
            alpha=CONFIG["point_alpha"],
            label="Test data",
        )

        # plot the regression plane
        ax.plot_surface(
            viz_data["xx"],
            viz_data["yy"],
            viz_data["zz"],
            alpha=CONFIG["plane_alpha"],
            color=CONFIG["plane_color"],
            rstride=2,
            cstride=2,
        )

        # add labels and title
        ax.set_xlabel("Square Footage")
        ax.set_ylabel("Bedrooms")
        ax.set_zlabel("Price (thousands $)")
        ax.set_title(
            "Multiple Linear Regression: 3D visualization with Regression Plane"
        )
        ax.legend()

        # add formula as text
        plt.figtext(0.1, 0.01, viz_data["formula_text"], fontsize=12)

        # save the figure to a file
        plt.savefig(output_file, bbox_inches="tight")

        logger.info(f"3d plot saved as {output_file}")

        if show_plot:
            plt.show()

        plt.close()
