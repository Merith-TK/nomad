#!/usr/bin/env python3

from __future__ import annotations

import argparse
import math
import sys
from pathlib import Path

BACKEND = None

try:
    from OCP.BRepMesh import BRepMesh_IncrementalMesh
    from OCP.IFSelect import IFSelect_RetDone
    from OCP.STEPControl import STEPControl_Reader
    from OCP.StlAPI import StlAPI_Writer

    BACKEND = "OCP"
except ModuleNotFoundError:
    try:
        from OCC.Core.BRepMesh import BRepMesh_IncrementalMesh
        from OCC.Core.IFSelect import IFSelect_RetDone
        from OCC.Core.STEPControl import STEPControl_Reader
        from OCC.Core.StlAPI import StlAPI_Writer

        BACKEND = "OCC"
    except ModuleNotFoundError:
        BACKEND = None


def convert_step_to_stl(
    input_path: Path,
    output_path: Path,
    linear_deflection: float,
    angular_deflection_deg: float,
    relative: bool,
    parallel: bool,
    ascii_mode: bool,
) -> None:
    reader = STEPControl_Reader()
    status = reader.ReadFile(str(input_path))
    if status != IFSelect_RetDone:
        raise RuntimeError(f"Failed to read STEP file: {input_path}")

    transferred = reader.TransferRoots()
    if transferred == 0:
        raise RuntimeError(f"No transferable geometry found in: {input_path}")

    shape = reader.OneShape()
    if shape.IsNull():
        raise RuntimeError(f"Loaded shape is null: {input_path}")

    angular_deflection_rad = math.radians(angular_deflection_deg)
    mesher = BRepMesh_IncrementalMesh(
        shape,
        linear_deflection,
        relative,
        angular_deflection_rad,
        parallel,
    )
    mesher.Perform()
    if not mesher.IsDone():
        raise RuntimeError("Meshing failed")

    output_path.parent.mkdir(parents=True, exist_ok=True)

    writer = StlAPI_Writer()
    if hasattr(writer, "SetASCIIMode"):
        writer.SetASCIIMode(ascii_mode)
    elif hasattr(writer, "ASCIIMode"):
        writer.ASCIIMode = ascii_mode
    writer.Write(shape, str(output_path))

    if not output_path.exists() or output_path.stat().st_size == 0:
        raise RuntimeError(f"Failed to write STL file: {output_path}")


def build_parser() -> argparse.ArgumentParser:
    parser = argparse.ArgumentParser(description="Convert STEP (.step/.stp) to STL")
    parser.add_argument("input", type=Path, help="Input STEP file path")
    parser.add_argument("output", type=Path, help="Output STL file path")
    parser.add_argument(
        "--linear-deflection",
        type=float,
        default=0.1,
        help="Linear deflection for meshing (smaller = finer mesh, default: 0.1)",
    )
    parser.add_argument(
        "--angular-deflection-deg",
        type=float,
        default=10.0,
        help="Angular deflection in degrees (smaller = finer mesh, default: 10)",
    )
    parser.add_argument(
        "--relative",
        action="store_true",
        help="Use relative deflection mode",
    )
    parser.add_argument(
        "--no-parallel",
        action="store_true",
        help="Disable parallel meshing",
    )
    parser.add_argument(
        "--ascii",
        action="store_true",
        help="Write ASCII STL instead of binary STL",
    )
    return parser


def main() -> int:
    parser = build_parser()
    args = parser.parse_args()

    if BACKEND is None:
        print(
            "Missing OpenCascade Python bindings. Install one of:\n"
            "  pip install OCP\n"
            "  pip install pythonocc-core",
            file=sys.stderr,
        )
        return 2

    input_path = args.input
    output_path = args.output

    if not input_path.exists():
        print(f"Input file does not exist: {input_path}", file=sys.stderr)
        return 2

    if input_path.suffix.lower() not in {".stp", ".step"}:
        print("Input file must have .stp or .step extension", file=sys.stderr)
        return 2

    try:
        convert_step_to_stl(
            input_path=input_path,
            output_path=output_path,
            linear_deflection=args.linear_deflection,
            angular_deflection_deg=args.angular_deflection_deg,
            relative=args.relative,
            parallel=not args.no_parallel,
            ascii_mode=args.ascii,
        )
    except Exception as exc:
        print(f"Conversion failed: {exc}", file=sys.stderr)
        return 1

    print(f"Wrote STL: {output_path} (backend: {BACKEND})")
    return 0


if __name__ == "__main__":
    raise SystemExit(main())
