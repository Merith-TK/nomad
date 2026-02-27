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
use <brain_mount_slim.scad>

// ============================================
// BRAIN MOUNT CONFIGURATION
// ============================================
// All brain mount settings consolidated here for easy adjustment:
// - Wedge 2 dimensions
// - Component positioning

// ============================================
// WEDGE 2 DIMENSIONS (Front wedge)
// ============================================
wedge2_size = [40.25, 20.0, 7.0];  // [Width, Length, Height] (mm)
wedge2_pos = [-31.25, 0, 0];       // [X, Y, Z] position offset (mm)

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
        // Wedge 1 from slim model (same origin/placement)
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