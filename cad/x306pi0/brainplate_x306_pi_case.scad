/*
 * Nomad Core - X306 + Pi Case Brainplate
 * Standalone brainplate with recessed cutout pocket
 */

// ============================================
// BRAINPLATE CONFIGURATION
// ============================================
// Outer block dimensions (adjusted to satisfy perimeter constraints)
base_width = 58.75;     // X dimension (mm) -> gives 3.5 mm + 2.0 mm side borders around 53.25 mm cutout
base_length = 85.0;     // Y dimension (mm)
base_height = 4.8;      // Z dimension (mm)

// Cutout block dimensions
cutout_width = 53.25;   // X dimension (mm)
cutout_length = 79.5;   // Y dimension (mm)
cutout_height = 40.0;   // Tall enough to fully pass through above cut floor (mm)

// Cutout placement
cutout_floor_z = 1.0;   // Leaves a 1.0 mm floor thickness in base block
cutout_border_x_left = 3.5; // Border on one long side wall (mm)

// Derived placement
cutout_x = cutout_border_x_left;                         // Leaves 2.0 mm on opposite long side wall
cutout_y = (base_length - cutout_length) / 2;           // Symmetric short-end borders (>= 2 mm)

// ============================================
// RENDER QUALITY
// ============================================
$fn = 64;

// ============================================
// MODEL
// ============================================
module x306_pi_brainplate() {
    difference() {
        cube([base_width, base_length, base_height], center=false);
        translate([cutout_x, cutout_y, cutout_floor_z])
            cube([cutout_width, cutout_length, cutout_height], center=false);
    }
}

// ============================================
// MAIN RENDER
// ============================================
x306_pi_brainplate();
