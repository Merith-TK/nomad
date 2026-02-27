/*
 * Nomad Core - Initial Baseplate v1.0
 * Connects puck mounting system to Raspberry Pi 3B+
 * 
 * Features:
 * - 25×25mm support column with rounded edges (7mm clearance above puck)
 * - 60×70mm baseplate with Pi 3B+ mounting holes
 * - Offset 7.5mm towards back/rounded edge of puck for balanced positioning
 */

// ============================================
// IMPORTS
// ============================================
use <parts/screw_cutout.scad>
use <parts/support_column.scad>

// ============================================
// RASPBERRY PI 3B+ DIMENSIONS
// ============================================
pi_width = 60.0;   // Baseplate width (X dimension, left-right)
pi_length = 70.0;  // Baseplate length (Y dimension, front-back)

// ============================================
// MOUNTING HOLE POSITIONS
// ============================================
// Hole edge offsets (measured from baseplate edges)
hole_offset_from_edge_x = 5.5;  // Distance from left/right edges (mm)
hole_offset_from_edge_y = 6.0;  // Distance from front/back edges (mm)

// ============================================
// SUPPORT COLUMN SPECIFICATIONS
// ============================================
support_size = 25.0;            // 25×25mm footprint on puck top (mm)
support_height = 7.0;           // Clearance above puck (mm)
support_thickness = 3.0;        // Wall thickness (mm)
support_corner_radius = 3.0;    // Rounded edge radius (mm)

// ============================================
// BASEPLATE SPECIFICATIONS
// ============================================
baseplate_width = pi_width;      // 60mm (mm)
baseplate_length = pi_length;    // 70mm (mm)
baseplate_thickness = 4.0;       // Platform thickness (mm)
baseplate_offset_y = 7.5;        // Offset towards puck back (mm)

// ============================================
// SCREW POSITIONING
// ============================================
screw_inset_from_bottom = -8.0;  // Hex nut recess inset from bottom (mm)

// ============================================
// RENDER QUALITY
// ============================================
$fn = 64;

// ============================================
// CALCULATED VALUES
// ============================================
// Hole spacing (center-to-center for reference)
pi_hole_spacing_x = 2 * (pi_width/2 - hole_offset_from_edge_x);   // 49mm (X/width)
pi_hole_spacing_y = 2 * (pi_length/2 - hole_offset_from_edge_y);  // 58mm (Y/length)

// Total assembly height
total_height = support_height + baseplate_thickness;  // 11mm

// ============================================
// BASEPLATE BODY
// ============================================
// Main platform for mounting Raspberry Pi

module baseplate_body() {
    // Platform is [width, length, thickness] = [X, Y, Z]
    cube([baseplate_width, baseplate_length, baseplate_thickness], center=true);
}

// ============================================
// RASPBERRY PI MOUNTING HOLES
// ============================================
// Four M2.5 screw holes matching Pi 3B+ pattern

module pi_mounting_holes() {
    // Calculate hole positions from baseplate center
    hole_pos_x = baseplate_width/2 - hole_offset_from_edge_x;   // 24.5mm from center
    hole_pos_y = baseplate_length/2 - hole_offset_from_edge_y;  // 29mm from center
    
    for (x = [-1, 1]) {
        for (y = [-1, 1]) {
            translate([
                x * hole_pos_x,
                y * hole_pos_y,
                -baseplate_thickness/2 + screw_inset_from_bottom
            ])
                screw_cutout();
        }
    }
}

// ============================================
// COMPLETE BASEPLATE ASSEMBLY
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
            
            // Orientation marker (small notch on front edge)
            translate([0, baseplate_offset_y + baseplate_length/2 - 2, support_height + baseplate_thickness - 0.5])
                rotate([0, 45, 0])
                cube([1, 3, 1], center=true);
        }
        
        // Cut mounting holes from baseplate
        translate([0, baseplate_offset_y, support_height + baseplate_thickness/2])
            pi_mounting_holes();
    }
}

// ============================================
// MAIN RENDER
// ============================================
baseplate_initial();