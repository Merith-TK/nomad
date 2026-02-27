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
// BRAIN MOUNT CONFIGURATION
// ============================================
// All brain mount settings consolidated here for easy adjustment:
// - Wedge dimensions (2 wedges with different sizes)
// - Screw hole positions (matching Pi 3B+ pattern)
// - Component positioning

// ============================================
// WEDGE 1 DIMENSIONS (Long wedge with screw holes)
// ============================================
wedge1_size = [20.0, 83.0, 7.0];  // [Width, Length, Height] (mm)

// ============================================
// WEDGE 2 DIMENSIONS (Front wedge)
// ============================================
wedge2_size = [40.25, 20.0, 7.0];  // [Width, Length, Height] (mm)
wedge2_pos = [-31.25, 0, 0];       // [X, Y, Z] position offset (mm)

// ============================================
// SCREW HOLE POSITIONS
// ============================================
// Hole positions on Wedge 1 match Pi 3B+ mounting pattern
hole1_offset = [6.5, 8.5];   // [Long edge, Short edge] distance (mm)
pi_length_spacing = 58.0;    // Standard Pi mounting spacing (mm)
hole2_offset = [hole1_offset[0] + pi_length_spacing, hole1_offset[1]];  // [64.5, 8.5] (mm)
screw_depth = -3.0;          // Inset from wedge surface (mm)

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
        wedge(wedge1_size[0], wedge1_size[1], wedge1_size[2]);
        
        // First screw hole
        translate([-hole1_offset[0], hole1_offset[1], screw_depth])
            screw_cutout();
        
        // Second screw hole (Pi lengthwise spacing)
        translate([-hole2_offset[0], hole2_offset[1], screw_depth])
            screw_cutout();
    }
}

// ============================================
// WEDGE 2 BODY
// ============================================
// Front wedge without screw holes

module wedge2_body() {
    wedge(wedge2_size[0], wedge2_size[1], wedge2_size[2]);
}

// ============================================
// COMPLETE BRAIN MOUNT ASSEMBLY
// ============================================
module brain_mount() {
    union() {
        // Wedge 1 (long wedge with screw holes)
        wedge1_with_holes();
        
        // Wedge 2 (positioned between screw holes)
        translate(wedge2_pos)
            wedge2_body();
    }
}

// ============================================
// MAIN RENDER
// ============================================
brain_mount();