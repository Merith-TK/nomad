/*
 * Nomad Core - Initial Baseplate v1.0
 * Connects puck to Raspberry Pi 3B+
 * 
 * Design: 25x25mm support structure with rounded edges, 6.5mm clearance
 * Baseplate: 60mm wide x 70mm long, offset 7.5mm towards puck back (rounded edge)
 */

// Import reusable modules
use <parts/screw_cutout.scad>
use <parts/support_column.scad>

// ============================================
// BASEPLATE DIMENSIONS
// ============================================
// Raspberry Pi 3B+ dimensions and mounting
pi_width = 60.0;           // Baseplate width (X dimension, left-right)
pi_length = 70.0;          // Baseplate length (Y dimension, front-back)

// Hole edge offsets (measured from edges)
hole_offset_from_edge_x = 5.5;  // Distance from left/right edges (width/X direction)
hole_offset_from_edge_y = 6.0;  // Distance from front/back edges (length/Y direction)

// Calculated hole spacing (for reference)
pi_hole_spacing_x = 2 * (pi_width/2 - hole_offset_from_edge_x);   // 60-5.5-5.5 = 49mm (X/width)
pi_hole_spacing_y = 2 * (pi_length/2 - hole_offset_from_edge_y);  // 70-6-6 = 58mm (Y/length)

// Support structure (puck to baseplate)
support_size = 25.0;       // 25x25mm footprint on puck top
support_height = 7.0;      // Clearance above puck
support_thickness = 3.0;   // Wall thickness of support structure
support_corner_radius = 3.0;  // Rounded edges on support column

// Baseplate positioning (offset towards back/rounded edge of puck)
baseplate_offset_y = 7.5;  // Offset towards back (5-10mm range)

// Baseplate 
baseplate_width = pi_width;
baseplate_length = pi_length;
baseplate_thickness = 4.0; // Thickness of main platform

// Screw cutout positioning
screw_inset_from_bottom = -8.0;  // Hex nut recess inset from bottom

// ============================================
// RENDER QUALITY
// ============================================
$fn = 64;

// ============================================
// MAIN BASEPLATE
// ============================================
module baseplate_body() {
    // Main platform
    // Cube is [X, Y, Z] where X=width (left-right), Y=length (front-back)
    cube([baseplate_width, baseplate_length, baseplate_thickness], center=true);
}

// ============================================
// RASPBERRY PI MOUNTING HOLES (Pi 3B+ spacing)
// ============================================
module pi_mounting_holes() {
    // Calculate hole positions from baseplate center
    // Baseplate cube is [width, length, thickness] = [X, Y, Z]
    // So X dimension = width (60mm), Y dimension = length (70mm)
    
    // Holes are 5.5mm from left/right edges (width/X direction)
    hole_pos_x = baseplate_width/2 - hole_offset_from_edge_x;   // 60/2 - 5.5 = 24.5mm
    
    // Holes are 6mm from front/back edges (length/Y direction)  
    hole_pos_y = baseplate_length/2 - hole_offset_from_edge_y;  // 70/2 - 6 = 29mm
    
    for (x = [-1, 1]) {
        for (y = [-1, 1]) {
            translate([
                x * hole_pos_x,  // ±24.5mm in X (width direction)
                y * hole_pos_y,  // ±29mm in Y (length direction)
                -baseplate_thickness/2 + screw_inset_from_bottom
            ])
                screw_cutout();
        }
    }
}

// ============================================
// MAIN ASSEMBLY MODULE
// ============================================
module baseplate_initial() {
    difference() {
        union() {
            // Support column (centered at origin, bottom at z=0)
            translate([0, 0, support_height/2])
                support_column(support_size, support_height, support_corner_radius);
            
            // Baseplate platform (offset towards back/rounded edge of puck)
            translate([0, baseplate_offset_y, support_height + baseplate_thickness/2])
                baseplate_body();
            
            // Alignment marker (small notch for orientation on front edge)
            translate([0, baseplate_offset_y + baseplate_length/2 - 2, support_height + baseplate_thickness - 0.5])
                rotate([0, 45, 0])
                cube([1, 3, 1], center=true);
        }
        
        // Cut mounting holes from top (with baseplate offset)
        translate([0, baseplate_offset_y, support_height + baseplate_thickness/2])
            pi_mounting_holes();
    }
}

// ============================================
// RENDER (when file is opened directly)
// ============================================
baseplate_initial();

// ============================================
// INFO OUTPUT
// ============================================
echo("=== INITIAL BASEPLATE v1.0 - Pi 3B+ ===");
echo(str("Baseplate: ", baseplate_width, "mm (W) x ", baseplate_length, "mm (L) x ", baseplate_thickness, "mm (H)"));
echo(str("Support structure: ", support_size, " x ", support_size, " x ", support_height, " mm (rounded corners)"));
echo(str("Baseplate offset (towards puck back): ", baseplate_offset_y, " mm (Y direction)"));
echo(str("Total height: ", support_height + baseplate_thickness, " mm"));
echo(str("Pi 3B+ mounting hole spacing: ", pi_hole_spacing_x, "mm (X) x ", pi_hole_spacing_y, "mm (Y)"));
echo(str("Hole edge offsets: ", hole_offset_from_edge_x, "mm (left/right/X), ", hole_offset_from_edge_y, "mm (front/back/Y)"));
echo("");
echo("Screw spec: M2.5 x 16mm with TR8 hex head");
echo(str("Hex nut recess inset: ", screw_inset_from_bottom, " mm from bottom"));
echo("Orientation: Offset towards rounded edge (back) of puck");
