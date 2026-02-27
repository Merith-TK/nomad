/*
 * Nomad Core - Screw Cutout Module
 * M2.5 screw hole with hex nut recess access from bottom
 * 
 * Screw specs: M2.5 threading, 16mm long, TR8 hex head
 */

// ============================================
// MODULE: SCREW CUTOUT
// ============================================
module screw_cutout() {
    union() {
        // Screw shaft hole (3mm diameter, 25mm tall)
        // Full height to allow screw to pass through
        cylinder(h=25, d=3, center=false);
        
        // Hex nut recess (8mm diameter, 6mm tall)
        // Starts 4.5mm from bottom to create nut pocket
        translate([0, 0, 4.5])
            cylinder(h=6, d=8, center=false);
    }
}

// ============================================
// RENDER QUALITY
// ============================================
$fn = 32;

// ============================================
// TEST RENDER (comment out when using as module)
// ============================================
// Render the cutout for visualization
screw_cutout();

// Show a reference cube to see the cutout in context
%translate([0, 0, 5])
    cube([15, 15, 5], center=true);

// ============================================
// INFO OUTPUT
// ============================================
echo("=== SCREW CUTOUT SPECIFICATIONS ===");
echo("Screw standard: M2.5");
echo("Screw length: 16mm");
echo("Hex head: TR8");
echo("");
echo("Shaft hole: 3mm diameter x 25mm height");
echo("Nut recess: 8mm diameter x 6mm height (starts at 4.5mm)");
echo("");
echo("Usage: difference() { your_part(); screw_cutout(); }");
