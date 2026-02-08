import sys
import argparse

from config import CONFIG
from housing_analysis import (
    DataProcessingError,
    ModelOperationError,
    evaluate_model,
    load_data,
    load_model,
    make_predictions,
    prepare_model_data,
    preprocess_data,
    save_model,
    train_model,
    print_results,
    create_visualization_data,
    create_2d_visualization,
    create_3d_visualization,
)

from housing_analysis import logger


def main() -> int:
    try:
        # parse command line args
        args = parse_arguments()

        # choose operation mode based on arguments
        if args.load_model:
            # load model and metadata from files
            model, scaler, metadata = load_model(
                args.model_path, args.metadata_path
            )

            if args.predict_only:
                df = load_data(args.file)
                processed_df = preprocess_data(df)

                predictions = make_predictions(processed_df, model, scaler)

                print(f"Length of predictions: {len(predictions)}")
            else:
                df = load_data(args.file)
                processed_df = preprocess_data(df)
                model_data = prepare_model_data(processed_df)
                model_results = evaluate_model(model_data, model, scaler)

                print(f"Test R2: {model_results.test_r2:.4f}")
        else:
            df = load_data(args.file)
            processed_df = preprocess_data(df)
            model_data = prepare_model_data(processed_df)

            model, scaler = train_model(model_data)

            model_results = evaluate_model(model_data, model, scaler)

            # print results to terminal
            print_results(model_data, model_results)

            # create visualization
            viz_data = create_visualization_data(model_data, model_results)
            show_plot = not args.no_plot

            create_2d_visualization(
                model_data,
                model_results,
                viz_data,
                CONFIG["output_image"],
                show_plot,
            )

            create_3d_visualization(
                model_data,
                model_results,
                viz_data,
                CONFIG['output_3d_image'],
                show_plot,
            )

            if args.save_model:
                save_model(model_results, args.model_path, args.metadata_path)

        return 0

    except DataProcessingError as e:
        logger.error(f"Data processing error: {str(e)}")
        return 1
    except ModelOperationError as e:
        logger.error(f"Model operation error: {str(e)}")
        return 2
    except Exception as e:
        logger.error(f"Unexpected error: {str(e)}")
        return 3


def parse_arguments() -> argparse.Namespace:
    parser = argparse.ArgumentParser(
        description="Multiple linear regression analysis on housing data",
    )

    parser.add_argument(
        "-f",
        "--file",
        type=str,
        default=CONFIG["default_csv"],
        help=f"Path to CSV file (default: {CONFIG['default_csv']})",
    )

    parser.add_argument(
        "--no-plot",
        action="store_true",
        help="Do not display the plot (still saves to file)",
    )

    parser.add_argument(
        "--save-model",
        action="store_true",
        help="Save the trained model to a file",
    )

    parser.add_argument(
        "--model-path",
        type=str,
        default=CONFIG["default_model_path"],
        help=f"Path to load/save model (default: {CONFIG['default_model_path']})",
    )

    parser.add_argument(
        "--load-model",
        action="store_true",
        help="Load a previously trained model instead of training a new one",
    )

    parser.add_argument(
        "--metadata-path",
        type=str,
        default=CONFIG["default_metadata_path"],
        help=f"Path to save/load metadata (default: {CONFIG['default_metadata_path']})",
    )

    parser.add_argument(
        "--predict-only",
        action="store_true",
        help="Only make predictions using a loaded model, no training or evaluation",
    )

    args, unknown = parser.parse_known_args()
    if unknown:
        logger.warning(f"Unknown arguments ignored: {unknown}")

    if args.predict_only and not args.load_model:
        parser.error("--predict-only requires --load-model")

    return args


if __name__ == "__main__":
    sys.exit(main())
