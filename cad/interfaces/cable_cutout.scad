/*
 * Nomad Core - Cable Cutout v0.1
 * Composite cable cutout shape for Stream Deck interface plate
 *
 * Build this cutout as a standalone module, then import with `use` into
 * streamdeck_interface_plate.scad for boolean subtraction.
 */

// ============================================
// CABLE CUTOUT CONFIGURATION
// ============================================
// Composite shape components are defined incrementally.

// Cuboid 1
cuboid_1_pos = [0, 0, 0];           // [X, Y, Z] (mm)
cuboid_1_rot = [0, 0, 0];           // [X, Y, Z] rotation (degrees)
cuboid_1_size = [20.0, 45.1, 16.5]; // [X, Y, Z] (mm)

// Cylinders (2x, identical except placement)
cylinder_dia = 16.5;                // Diameter (mm)
cylinder_length = 20.0;             // Length along X (mm)

// Cuboid 2 (subtractive boolean)
cuboid_2_size = [20.0, 75.0, 6.0];  // [X, Y, Z] (mm)
cuboid_2_pos = [0.0, -15.0, 0.0];   // [X, Y, Z] (mm) - cuts lower 6mm across shape

// ============================================
// RENDER QUALITY
// ============================================
preview_fn = 20;   // Faster viewport interaction
export_fn = 48;    // Better quality for STL export
$fn = $preview ? preview_fn : export_fn;

// ============================================
// COMPOSITE CABLE CUTOUT
// ============================================
module cable_cutout() {
    difference() {
        union() {
            translate(cuboid_1_pos)
                rotate(cuboid_1_rot)
                    cube(cuboid_1_size, center=false);

            translate([
                cuboid_1_pos[0],
                cuboid_1_pos[1],
                cuboid_1_pos[2] + cuboid_1_size[2] / 2
            ])
                rotate([0, 90, 0])
                    cylinder(h=cylinder_length, d=cylinder_dia, center=false);

            translate([
                cuboid_1_pos[0],
                cuboid_1_pos[1] + cuboid_1_size[1],
                cuboid_1_pos[2] + cuboid_1_size[2] / 2
            ])
                rotate([0, 90, 0])
                    cylinder(h=cylinder_length, d=cylinder_dia, center=false);
        }

        translate(cuboid_2_pos)
            cube(cuboid_2_size, center=false);
    }
}

// ============================================
// MAIN RENDER
// ============================================
cable_cutout();
