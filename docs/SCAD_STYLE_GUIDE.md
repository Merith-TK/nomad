# OpenSCAD Style Guide for NOMAD Core

This document defines coding standards for all `.scad` files in the NOMAD Core project to ensure consistency, readability, and maintainability.

## File Structure

Every `.scad` file should follow this template:

```scad
/*
 * Nomad Core - [Component Name]
 * [Brief description of what this component does]
 * 
 * [Optional: Additional details, design notes, or specifications]
 */

// ============================================
// IMPORTS
// ============================================
use <parts/module_name.scad>

// ============================================
// [SECTION NAME]
// ============================================
// Parameters with inline comments
variable_name = value;  // Description with units

// ============================================
// RENDER QUALITY
// ============================================
$fn = 64;  // Or 32 for simple parts

// ============================================
// MODULE DEFINITIONS
// ============================================
module component_name() {
    // Implementation
}

// ============================================
// MAIN RENDER
// ============================================
component_name();
```

## Naming Conventions

### Variables
- **Format**: `snake_case` (all lowercase with underscores)
- **Be descriptive**: `wedge1_width` not `w1`
- **Include units in comments**: `base_height = 3.5;  // mm`
- **Group related variables**: Keep dimensions together, screw specs together, etc.

**Examples:**
```scad
puck_size = 33.5;           // Square footprint (mm)
base_height = 3.5;          // Base thickness (mm)
mounting_hole_dia = 3.5;    // Mounting hole diameter (mm)
```

### Modules
- **Format**: `snake_case`
- **Reusable modules**: Place in `cad/parts/` directory
- **Main components**: Place in `cad/` root
- **Be specific**: `screw_cutout()` not `hole()`

**Examples:**
```scad
module support_column(size, height, corner_radius) { }
module screw_cutout() { }
module puck_body_with_selective_rounds() { }
```

### Files
- **Format**: `snake_case.scad`
- **Descriptive names**: `brain_mount.scad` not `mount.scad`
- **Assembly files**: Prefix with number for ordering: `00_assembly_complete.scad`
- **Archive files**: Keep version in name: `puck_v2.1.scad`

## Section Organization

Use divider comments to separate logical sections:

```scad
// ============================================
// SECTION NAME (ALL CAPS)
// ============================================
```

### Standard Section Order

1. **File header comment** (/* ... */)
2. **IMPORTS** - `use <>` and `include <>` statements
3. **MAIN DIMENSIONS** - Primary size parameters
4. **FEATURE SECTIONS** - Group related parameters (mounting holes, corner rounding, etc.)
5. **RENDER QUALITY** - `$fn` setting
6. **CALCULATED VALUES** - Derived from parameters (optional but recommended)
7. **MODULE DEFINITIONS** - All module declarations
8. **MAIN RENDER** - The actual component render call

## Comments

### Parameter Comments
- **Inline format**: Variable assignment on left, `// Comment` on right
- **Include units**: Always specify mm, degrees, etc.
- **Explain "why"**: Not just "what"

```scad
// GOOD
hole1_offset_long = 6.5;   // Distance from long edge (83mm edge)
front_corner_radius = 8.0; // TIGHT fit (tested: 7.5=loose, 8.0=tight)

// ACCEPTABLE
wedge1_width = 20;  // Triangle-to-triangle (long width)

// AVOID
x = 20;  // Width
```

### Module Documentation
Document module parameters using block comments:

```scad
// ============================================
// SUPPORT COLUMN MODULE
// ============================================
// Parameters:
// - size: footprint size (mm)
// - height: column height (mm)
// - corner_radius: rounded corner radius (mm)
module support_column(size=25, height=7, corner_radius=3) {
```

### Implementation Comments
Use single-line comments to explain complex logic:

```scad
// Create body with different corner radii
// Front (positive Y) = rounded for insertion
// Back (negative Y) = sharper corners
```

## Parameters and Customization

### Grouping
Group related parameters under clear section headers:

```scad
// ============================================
// WEDGE 1 DIMENSIONS
// ============================================
wedge1_width = 20;
wedge1_length = 83.0;
wedge1_height = 7.0;

// ============================================
// WEDGE 2 DIMENSIONS
// ============================================
wedge2_width = 40.25;
wedge2_length = 20.0;
wedge2_height = 7.0;
```

### Calculated Values
Show derivations for calculated values:

```scad
// ============================================
// CALCULATED VALUES
// ============================================
// Hole spacing derived from board edge offsets
pi_hole_spacing_x = 2 * (pi_width/2 - hole_offset_from_edge_x);   // 60-5.5-5.5 = 49mm
pi_hole_spacing_y = 2 * (pi_length/2 - hole_offset_from_edge_y);  // 70-6-6 = 58mm

// Second screw hole uses Pi 3B+ lengthwise spacing
hole2_offset_long = hole1_offset_long + pi_length_spacing;  // 6.5 + 58 = 64.5mm
```

### Default Values
Provide sensible defaults for module parameters:

```scad
module support_column(size=25, height=7, corner_radius=3) {
    // Defaults match most common use case
}
```

## Code Formatting

### Indentation
- **Use 4 spaces** (not tabs)
- Indent once per nesting level

```scad
module example() {
    difference() {
        cube([10, 10, 10]);
        translate([5, 5, 0])
            cylinder(h=15, d=3);
    }
}
```

### Line Length
- **Keep lines under 100 characters** when possible
- Break long parameter lists across multiple lines:

```scad
// GOOD
translate([
    x * (size/2 - corner_radius),
    y * (size/2 - corner_radius),
    0
])
    cylinder(h=height, r=corner_radius);

// AVOID
translate([x * (size/2 - corner_radius), y * (size/2 - corner_radius), 0]) cylinder(h=height, r=corner_radius);
```

### Alignment
Align related values for readability:

```scad
// Values aligned for easy scanning
wedge1_width  = 20.0;
wedge1_length = 83.0;
wedge1_height = 7.0;

// Or group logically
hole1_offset_short = 8.5;   // Distance from short edge
hole1_offset_long  = 6.5;   // Distance from long edge
```

## Render Quality

### `$fn` Settings
Set `$fn` explicitly in every file:

```scad
// ============================================
// RENDER QUALITY
// ============================================
$fn = 64;  // High quality for complex curves

// OR

$fn = 32;  // Lower quality for simple parts (faster)
```

**Guidelines:**
- `$fn = 64` - Default for most components
- `$fn = 32` - Simple parts, faster preview
- `$fn = 128` - Final export quality (optional)

## Module Design

### Reusable Modules
Place shared modules in `cad/parts/`:

```scad
// In cad/parts/screw_cutout.scad
module screw_cutout() {
    // Reusable across multiple components
}

// In cad/brain_mount.scad
use <parts/screw_cutout.scad>
```

### Module Parameters
- Order parameters from most to least important
- Provide defaults for optional parameters
- Document parameter units and purpose

```scad
// Well-documented, parameterized module
module wedge(width, length, height) {
    // width  = triangle-to-triangle dimension (mm)
    // length = back wall to nose dimension (mm)
    // height = tall dimension (mm)
    
    linear_extrude(height=height)
        polygon([
            [0, 0],
            [width, 0],
            [0, length]
        ]);
}
```

### Design for Customization
Make components easy to modify:

```scad
// GOOD - Easy to customize
base_height = 3.5;           // Base thickness (mm)
corner_radius = 8.0;         // TIGHT fit (tested: 7.5=loose, 8.0=tight)
tolerance = 0.0;             // Adjust for printer calibration

// AVOID - Magic numbers
cube([33.5, 33.5, 3.5]);
```

## Testing and Development

### Test Renders
Include commented-out test renders at the bottom of module files:

```scad
// ============================================
// TEST RENDER (comment out when using as module)
// ============================================
// screw_cutout();
```

### Version History
Keep old versions in `cad/archive/` with version numbers:
- `puck_v2.1.scad` (current)
- `puck_v2.0.scad` (archive)
- `puck_v1.0.scad` (archive)

## Best Practices

### Do's
✅ Use descriptive variable names  
✅ Include units in comments  
✅ Group related parameters  
✅ Document module parameters  
✅ Show calculation derivations  
✅ Set `$fn` explicitly  
✅ Use section dividers  
✅ Keep modules focused and reusable  

### Don'ts
❌ Use single-letter variable names  
❌ Mix units (stick to mm)  
❌ Use magic numbers without explanation  
❌ Create monolithic files (extract reusable modules)  
❌ Leave `$fn` at default  
❌ Skip comments on complex calculations  

## Example: Complete File

```scad
/*
 * Nomad Core - Example Component
 * Demonstrates proper style guide adherence
 * 
 * This component shows all the standard formatting conventions
 */

// ============================================
// IMPORTS
// ============================================
use <parts/screw_cutout.scad>

// ============================================
// MAIN DIMENSIONS
// ============================================
component_width = 50.0;   // Overall width (mm)
component_length = 60.0;  // Overall length (mm)
component_height = 5.0;   // Base thickness (mm)

// ============================================
// MOUNTING FEATURES
// ============================================
hole_diameter = 3.0;      // M2.5 screw clearance (mm)
hole_spacing = 40.0;      // Center-to-center spacing (mm)

// ============================================
// RENDER QUALITY
// ============================================
$fn = 64;

// ============================================
// CALCULATED VALUES
// ============================================
hole_offset = hole_spacing / 2;  // 40 / 2 = 20mm from center

// ============================================
// MAIN COMPONENT
// ============================================
module example_component() {
    difference() {
        // Main body
        cube([component_width, component_length, component_height]);
        
        // Mounting holes
        for (x = [-1, 1]) {
            translate([component_width/2 + x * hole_offset, component_length/2, 0])
                screw_cutout();
        }
    }
}

// ============================================
// MAIN RENDER
// ============================================
example_component();
```

## Enforcement

When reviewing `.scad` files:
1. Check header comment is present
2. Verify section dividers are used
3. Confirm variables use `snake_case`
4. Ensure parameters have unit comments
5. Validate `$fn` is set explicitly
6. Check for reusable modules in `cad/parts/`

This style guide is a living document. Update it as new patterns emerge or better practices are discovered.
