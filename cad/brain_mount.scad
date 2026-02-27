/*
 * Nomad Core - Brain Mount Bracket
 * Attaches to left side of main baseplate for mounting the "brain" module
 * 
 * Two wedge-shaped mounting brackets with screw holes
 */

// Import screw cutout module
use <parts/screw_cutout.scad>

// ============================================
// WEDGE DIMENSIONS
// ============================================
// Wedge 1 (long width wedge)
wedge1_width = 20;    // Triangle-to-triangle (long width)
wedge1_length = 83.0;   // Back wall to nose
wedge1_height = 7.0;    // Tall dimension

// Wedge 2 (front wedge)
wedge2_width = 40.25;    // Triangle-to-triangle
wedge2_length = 20.0;  // Back wall to nose
wedge2_height = 7.0;    // Tall dimension

// ============================================
// SCREW HOLE POSITIONS
// ============================================
// Screw holes are on wedge 1 (the long 83mm width wedge)
// First hole position (from left corner)
hole1_offset_short = 8.5;  // Distance from short edge (20mm edge)
hole1_offset_long = 6.5;   // Distance from long edge (83mm edge)

// Second hole spacing (along the long 83mm wedge)
// Uses Pi 3B+ lengthwise hole spacing
pi_length_spacing = 58.0;  // From baseplate_initial.scad
hole2_offset_long = hole1_offset_long + pi_length_spacing;  // 8.5 + 58 = 66.5mm

// ============================================
// RENDER QUALITY
// ============================================
$fn = 64;

// ============================================
// RIGHT ANGLE WEDGE MODULE
// ============================================
module wedge(width, length, height) {
    // Create a right-angle triangular prism
    // width = triangle-to-triangle dimension
    // length = back wall to nose dimension
    // height = tall dimension
    
    // Rotate counter-clockwise (90 degrees around Z axis)
    rotate([0, 0, 90])
    rotate([90, 0, 0])
    translate([0, 0, -length])
    linear_extrude(height=length)
        polygon([
            [0, 0],              // Origin
            [width, 0],          // Right edge
            [0, height]          // Top corner
        ]);
}

// ============================================
// WEDGE 1 WITH SCREW HOLES (long width wedge)
// ============================================
module wedge1_with_holes() {
    difference() {
        screw_depth = -3.0;
        
        // Main wedge body
        wedge(wedge1_width, wedge1_length, wedge1_height);
        
        // First screw hole
        translate([-hole1_offset_long, hole1_offset_short, screw_depth])
            screw_cutout();
        
        // Second screw hole (spaced by Pi lengthwise spacing)
        translate([-hole2_offset_long, hole1_offset_short, screw_depth])
            screw_cutout();
    }
}

// ============================================
// WEDGE 2 (no holes)
// ============================================
module wedge2_body() {
    wedge(wedge2_width, wedge2_length, wedge2_height);
}

// ============================================
// COMPLETE BRAIN MOUNT ASSEMBLY
// ============================================
module brain_mount() {
    union() {
        // Wedge 1 (long width wedge with screw holes)
        wedge1_with_holes();
        
        // Wedge 2 (positioned between the two screw holes)
        translate([-31.25, 0, 0])
            wedge2_body();
    }
}

// ============================================
// RENDER (when file is opened directly)
// ============================================
brain_mount();

// ============================================
// INFO OUTPUT
// ============================================
echo("=== BRAIN MOUNT BRACKET ===");
echo(str("Wedge 1 (long width): ", wedge1_width, "mm (W) x ", wedge1_length, "mm (L) x ", wedge1_height, "mm (H)"));
echo(str("Wedge 2 (front): ", wedge2_width, "mm (W) x ", wedge2_length, "mm (L) x ", wedge2_height, "mm (H)"));
echo("");
echo("Screw holes in Wedge 1 (long width wedge):");
echo(str("  Hole 1: ", hole1_offset_long, "mm from long edge, ", hole1_offset_short, "mm from short edge"));
echo(str("  Hole 2: ", hole2_offset_long, "mm from long edge, ", hole1_offset_short, "mm from short edge"));
echo(str("  Spacing: ", pi_length_spacing, "mm (matches Pi lengthwise spacing)"));
echo("");
echo("Screw spec: M2.5 x 16mm with TR8 hex head");
