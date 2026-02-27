/*
 * Nomad Core - Screw Cutout Module
 * M2.5 screw hole with hex nut recess access from bottom
 * 
 * Screw specs: M2.5 threading, 16mm long, TR8 hex head
 */

// ============================================
// SCREW SPECIFICATIONS
// ============================================
screw_shaft_diameter = 3.0;   // Clearance for M2.5 screw (mm)
screw_shaft_height = 25.0;    // Full height to pass through (mm)
nut_recess_diameter = 8.0;    // Hex nut pocket diameter (mm)
nut_recess_height = 6.0;      // Hex nut pocket depth (mm)
nut_recess_offset = 4.5;      // Start position from bottom (mm)

// ============================================
// RENDER QUALITY
// ============================================
$fn = 32;

// ============================================
// SCREW CUTOUT MODULE
// ============================================
// Creates M2.5 screw hole with hex nut recess accessible from bottom
//
// Parameters: None (uses global specifications above)
//
// Usage: difference() { your_part(); translate([x,y,z]) screw_cutout(); }

module screw_cutout() {
    union() {
        // Screw shaft hole - full height to allow screw to pass through
        cylinder(h=screw_shaft_height, d=screw_shaft_diameter, center=false);
        
        // Hex nut recess - starts offset from bottom to create nut pocket
        translate([0, 0, nut_recess_offset])
            cylinder(h=nut_recess_height, d=nut_recess_diameter, center=false);
    }
}

// ============================================
// TEST RENDER (uncomment for visualization)
// ============================================
// screw_cutout();
//
// // Show reference cube to see the cutout in context
// %translate([0, 0, 5])
//     cube([15, 15, 5], center=true);