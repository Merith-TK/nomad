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

// Wedge (additive shape)
// Upside down: sloped face on the underside, nose pointing toward +X
wedge_cutout_pos = [7.0, 78.5, 6.0];     // [X, Y, Z] (mm)
wedge_cutout_rot = [90.0, 5.54, 0.0];      // [X, Y, Z] rotation (degrees)
wedge_cutout_size = [60.0, 112, 15.0];   // [X, Y, Z] (mm)
wedge_edge_radius = 1.5;                 // Rounded edge radius on all wedge edges (mm)
rounding_fn = $preview ? 10 : 20;        // Sphere smoothness for full-edge rounding
show_wedge_cutout_preview = true;        // Show/hide wedge overlay for positioning

// ============================================
// RENDER QUALITY
// ============================================
preview_fn = 20;   // Faster viewport interaction
export_fn = 48;    // Better quality for STL export
$fn = $preview ? preview_fn : export_fn;

// ============================================
// COMPOSITE CABLE CUTOUT
// ============================================
module wedge_cutout_volume() {
    wedge_x = wedge_cutout_size[0];
    wedge_y = wedge_cutout_size[1];
    wedge_z = wedge_cutout_size[2];
    safe_wedge_radius = min(wedge_edge_radius, wedge_x / 2 - 0.01, wedge_y / 2 - 0.01, wedge_z / 2 - 0.01);
    core_x = wedge_x - (2 * safe_wedge_radius);
    core_y = wedge_y - (2 * safe_wedge_radius);
    core_z = wedge_z - (2 * safe_wedge_radius);

    translate(wedge_cutout_pos)
        rotate(wedge_cutout_rot)
            minkowski() {
                translate([safe_wedge_radius, safe_wedge_radius, safe_wedge_radius])
                    linear_extrude(height=core_y, center=false)
                        polygon([
                            [0, 0],
                            [0, core_z],
                            [core_x, core_z]
                        ]);

                sphere(r=safe_wedge_radius, $fn=rounding_fn);
            }
}

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

            wedge_cutout_volume();
        }

        translate(cuboid_2_pos)
            cube(cuboid_2_size, center=false);
    }
}

// ============================================
// MAIN RENDER
// ============================================
cable_cutout();

if ($preview && show_wedge_cutout_preview) {
    #color([1.0, 0.5, 0.1, 0.6])
        wedge_cutout_volume();
}
