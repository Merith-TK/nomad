# CAD Source Files

OpenSCAD parametric source files for 3D printable components.

## Active Designs

- **puck_design.scad** - Main puck module (tested fit: 8mm front corners)
- **puck_fit_test.scad** - Fit test set with tolerance variations
- **baseplate_initial.scad** - Puck-to-Pi Zero mounting plate with 6.5mm clearance
- **screw_cutout.scad** - Reusable M2.5 screw hole module with hex nut recess

## Assembly Files

- **assembly_complete.scad** - All-in-one view of full system (imports other files)
  - Toggle components with `show_puck` and `show_baseplate` variables
  - Set `show_exploded = true` for exploded view

See [README_IMPORTS.md](README_IMPORTS.md) for details on using `use` and `include` to import SCAD files.

## Archive

Old reference designs and converted files from original STLs.

## Workflow

1. Edit SCAD files in this directory
2. Render to STL: `openscad -o ../stl/latest/<name>.stl <file>.scad`
3. Test print and iterate
4. Archive old versions to `../stl/archive/` when making breaking changes
