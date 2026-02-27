# OpenSCAD File Imports

OpenSCAD supports importing other SCAD files to create modular, reusable designs.

## Two Import Methods

### `use <file.scad>`
Imports **modules and functions only** - doesn't auto-render anything.

```openscad
use <puck_design.scad>
use <baseplate_initial.scad>

// Now you can use the modules
translate([0, 0, 10])
    puck_design();

translate([20, 0, 0])
    baseplate_initial();
```

**Best for:** Assembly files where you want to position components yourself.

### `include <file.scad>`
Imports **and renders everything** from that file.

```openscad
include <puck_design.scad>  // Renders the puck immediately

// Add more stuff around it
translate([20, 0, 0])
    cube([10, 10, 10]);
```

**Best for:** Shared variables/constants, or when you want to render the imported file.

## Assembly Pattern

See [assembly_complete.scad](assembly_complete.scad) for a complete example:

```openscad
use <puck_design.scad>
use <baseplate_initial.scad>

// Position components
puck_design();  // At origin

translate([0, 0, 3.5])  // Stack baseplate on top
    baseplate_initial();
```

## Module Requirements

For a SCAD file to be importable with `use`, wrap your main code in a module:

**❌ Won't work with `use`:**
```openscad
// Top-level code renders immediately
cube([10, 10, 10]);
```

**✅ Works with `use`:**
```openscad
module my_part() {
    cube([10, 10, 10]);
}

// Render when opened directly
my_part();
```

When imported with `use`, only the `module my_part()` is available - the `my_part();` call at the bottom doesn't execute.

## Tips

- **File paths** are relative to the file doing the import
- **Circular imports** will cause errors
- **Shared constants** can go in a separate file and use `include`
- **Color/transparency** only affects preview mode, not final renders
