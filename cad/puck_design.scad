/*
 * Nomad Core - Puck v2.1 (With Insertion Guide Rounds)
 * Quick-release mounting interface with selective corner rounding
 * 
 * Front corners are rounded for easy sliding insertion into mount
 * Back corners are sharper for tight fit when fully inserted
 */

// ============================================
// MAIN DIMENSIONS
// ============================================
puck_size = 33.5;           // Square footprint (mm)
base_height = 3.5;          // Base thickness (mm)

// ============================================
// CORNER ROUNDING
// ============================================
// Front two corners (toward hand) are rounded for insertion guidance
front_corner_radius = 8.0;  // TIGHT fit (tested: 7.5=loose, 8.0=tight)
back_corner_radius = 1.0;   // Sharper corners for secure fit

// ============================================
// MOUNTING HOLES (OPTIONAL)
// ============================================
add_mounting_holes = false;      // Toggle mounting holes
mounting_hole_dia = 3.5;         // Hole diameter (mm)
mounting_hole_spacing = 25.0;    // Center-to-center spacing (mm)
mounting_hole_depth = 3.0;       // Depth from top (mm)

// ============================================
// FIT ADJUSTMENT
// ============================================
tolerance = 0.0;  // Adjust for printer calibration (mm)

// ============================================
// RENDER QUALITY
// ============================================
$fn = 64;

// ============================================
// CALCULATED VALUES
// ============================================
actual_size = puck_size + tolerance;
mounting_offset = mounting_hole_spacing / 2;

// ============================================
// PUCK BODY WITH SELECTIVE CORNER ROUNDING
// ============================================
// Creates base body with different corner radii
// Front (positive Y) = rounded for insertion
// Back (negative Y) = sharper corners for fit

module puck_body_with_selective_rounds() {
    hull() {
        // Front-right corner (rounded for insertion)
        translate([actual_size/2 - front_corner_radius, actual_size/2 - front_corner_radius, 0])
            cylinder(h=base_height, r=front_corner_radius);
        
        // Front-left corner (rounded for insertion)
        translate([-actual_size/2 + front_corner_radius, actual_size/2 - front_corner_radius, 0])
            cylinder(h=base_height, r=front_corner_radius);
        
        // Back-right corner (sharper)
        translate([actual_size/2 - back_corner_radius, -actual_size/2 + back_corner_radius, 0])
            cylinder(h=base_height, r=back_corner_radius);
        
        // Back-left corner (sharper)
        translate([-actual_size/2 + back_corner_radius, -actual_size/2 + back_corner_radius, 0])
            cylinder(h=base_height, r=back_corner_radius);
    }
}

// ============================================
// MOUNTING HOLES
// ============================================
// Optional mounting holes for securing puck to band mount

module mounting_holes() {
    if (add_mounting_holes) {
        hole_depth = (mounting_hole_depth > base_height) ? base_height + 0.2 : mounting_hole_depth + 0.1;
        
        for (x = [-1, 1]) {
            for (y = [-1, 1]) {
                translate([x * mounting_offset, y * mounting_offset, base_height - hole_depth + 0.1])
                    cylinder(h=hole_depth, d=mounting_hole_dia);
            }
        }
    }
}

// ============================================
// COMPLETE PUCK ASSEMBLY
// ============================================
module puck_design() {
    difference() {
        puck_body_with_selective_rounds();
        mounting_holes();
    }
    
    // Orientation marker (small notch on front edge)
    translate([0, actual_size/2 - 1, base_height/2])
        rotate([0, 45, 0])
        cube([1, 2, 1], center=true);
}

// ============================================
// MAIN RENDER
// ============================================
puck_design();