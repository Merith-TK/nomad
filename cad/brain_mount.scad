/*
 * Nomad Core - Brain Mount Bracket
 * Mounting bracket for compute module (Raspberry Pi or compatible)
 * 
 * Two wedge-shaped brackets with M2.5 screw holes for secure mounting
 * Designed to attach to left side of main baseplate
 */

// ============================================
// IMPORTS
// ============================================
use <parts/screw_cutout.scad>

// ============================================
// WEDGE 1 DIMENSIONS (Long wedge with screw holes)
// ============================================
wedge1_width = 20.0;    // Triangle-to-triangle width (mm)
wedge1_length = 83.0;   // Back wall to nose length (mm)
wedge1_height = 7.0;    // Tall dimension (mm)

// ============================================
// WEDGE 2 DIMENSIONS (Front wedge)
// ============================================
wedge2_width = 40.25;   // Triangle-to-triangle width (mm)
wedge2_length = 20.0;   // Back wall to nose length (mm)
wedge2_height = 7.0;    // Tall dimension (mm)

// ============================================
// SCREW HOLE POSITIONS
// ============================================
// Hole positions on Wedge 1 match Pi 3B+ mounting pattern
hole1_offset_short = 8.5;  // Distance from short edge (20mm edge)
hole1_offset_long = 6.5;   // Distance from long edge (83mm edge)

// Pi 3B+ lengthwise hole spacing for second hole
pi_length_spacing = 58.0;  // Standard Pi mounting spacing (mm)
hole2_offset_long = hole1_offset_long + pi_length_spacing;  // 6.5 + 58 = 64.5mm

// Screw inset depth
screw_depth = -3.0;  // Inset from wedge surface (mm)

// ============================================
// WEDGE 2 POSITION
// ============================================
wedge2_offset_x = -31.25;  // Position between screw holes (mm)

// ============================================
// RENDER QUALITY
// ============================================
$fn = 64;

// ============================================
// RIGHT ANGLE WEDGE MODULE
// ============================================
// Creates a right-angle triangular prism
//
// Parameters:
// - width: Triangle-to-triangle dimension (mm)
// - length: Back wall to nose dimension (mm)
// - height: Tall dimension (mm)

module wedge(width, length, height) {
    // Rotate to proper orientation for mounting
    rotate([0, 0, 90])
    rotate([90, 0, 0])
    translate([0, 0, -length])
    linear_extrude(height=length)
        polygon([
            [0, 0],        // Origin
            [width, 0],    // Right edge
            [0, height]    // Top corner
        ]);
}

// ============================================
// WEDGE 1 WITH SCREW HOLES
// ============================================
// Long wedge with two M2.5 screw holes matching Pi mounting pattern

module wedge1_with_holes() {
    difference() {
        // Main wedge body
        wedge(wedge1_width, wedge1_length, wedge1_height);
        
        // First screw hole
        translate([-hole1_offset_long, hole1_offset_short, screw_depth])
            screw_cutout();
        
        // Second screw hole (Pi lengthwise spacing)
        translate([-hole2_offset_long, hole1_offset_short, screw_depth])
            screw_cutout();
    }
}

// ============================================
// WEDGE 2 BODY
// ============================================
// Front wedge without screw holes

module wedge2_body() {
    wedge(wedge2_width, wedge2_length, wedge2_height);
}

// ============================================
// COMPLETE BRAIN MOUNT ASSEMBLY
// ============================================
module brain_mount() {
    union() {
        // Wedge 1 (long wedge with screw holes)
        wedge1_with_holes();
        
        // Wedge 2 (positioned between screw holes)
        translate([wedge2_offset_x, 0, 0])
            wedge2_body();
    }
}

// ============================================
// MAIN RENDER
// ============================================
brain_mount();