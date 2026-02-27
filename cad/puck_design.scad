/*
 * Nomad Core - Puck v2.1 (With Insertion Guide Rounds)
 * Two adjacent corners rounded for easy sliding insertion
 * 
 * SOLUTION: Rounded "leading edge" corners help guide the puck into mount
 */

// ============================================
// MAIN DIMENSIONS
// ============================================
puck_size = 33.5;           // Square footprint (mm)
base_height = 3.5;          // Base thickness (mm)

// ============================================
// CORNER ROUNDING - For easier insertion
// ============================================
// Front two corners (toward hand) are rounded for insertion
front_corner_radius = 8.0;  // Front corners - TIGHT fit (tested: 7.5=loose, 8.0=tight)
back_corner_radius = 1.0;   // Back corners - smaller radius (sharper)

// ============================================
// MOUNTING HOLES
// ============================================
add_mounting_holes = false;
mounting_hole_dia = 3.5;
mounting_hole_spacing = 25.0;
mounting_hole_depth = 3.0;

// ============================================
// FIT ADJUSTMENT
// ============================================
tolerance = 0.0;

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
// MAIN PUCK BODY WITH SELECTIVE CORNER ROUNDING
// ============================================
module puck_body_with_selective_rounds() {
    // Create body with different corner radii
    // Front (positive Y) = rounded for insertion
    // Back (negative Y) = sharper
    
    hull() {
        // Front-right corner (rounded)
        translate([actual_size/2 - front_corner_radius, actual_size/2 - front_corner_radius, 0])
            cylinder(h=base_height, r=front_corner_radius);
        
        // Front-left corner (rounded)
        translate([-actual_size/2 + front_corner_radius, actual_size/2 - front_corner_radius, 0])
            cylinder(h=base_height, r=front_corner_radius);
        
        // Back-right corner (sharp)
        translate([actual_size/2 - back_corner_radius, -actual_size/2 + back_corner_radius, 0])
            cylinder(h=base_height, r=back_corner_radius);
        
        // Back-left corner (sharp)
        translate([-actual_size/2 + back_corner_radius, -actual_size/2 + back_corner_radius, 0])
            cylinder(h=base_height, r=back_corner_radius);
    }
}

// ============================================
// MOUNTING HOLES
// ============================================
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
// COMPLETE PUCK MODULE
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
// RENDER (when file is opened directly)
// ============================================
puck_design();

// ============================================
// INFO OUTPUT
// ============================================
echo(str("=== PUCK v2.1 - INSERTION-FRIENDLY ==="));
echo(str("Size: ", actual_size, " x ", actual_size, " x ", base_height, " mm"));
echo(str("Front corners (rounded): ", front_corner_radius, " mm - EASY INSERTION"));
echo(str("Back corners (sharp): ", back_corner_radius, " mm"));
echo(str("Tolerance: ", tolerance, " mm"));
if (add_mounting_holes) {
    echo(str("Mounting holes: ", mounting_hole_dia, " mm @ ", mounting_hole_spacing, " mm spacing"));
}
echo("");
echo("Front = rounded for sliding in, Back = sharper for fit");
