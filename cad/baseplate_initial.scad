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
// BASEPLATE CONFIGURATION
// ============================================
// All baseplate settings consolidated here for easy adjustment:
// - Raspberry Pi dimensions and mounting
// - Support column specifications
// - Baseplate platform specifications
// - Screw positioning

// ============================================
// RASPBERRY PI 3B+ SPECIFICATIONS
// ============================================
pi_size = [60.0, 70.0];          // [Width, Length] (mm)
pi_hole_offsets = [5.5, 6.0];    // [X edge, Y edge] distance from edges (mm)

// ============================================
// SUPPORT COLUMN SPECIFICATIONS
// ============================================
support_size = 25.0;             // 25×25mm footprint on puck top (mm)
support_height = 7.0;            // Clearance above puck (mm)
support_corner_radius = 3.0;     // Rounded edge radius (mm)

// ============================================
// BASEPLATE PLATFORM SPECIFICATIONS
// ============================================
baseplate_size = pi_size;        // [60, 70] matches Pi dimensions (mm)
baseplate_thickness = 4.0;       // Platform thickness (mm)
baseplate_corner_radius = 2.0;   // Outer corner radius for safer handling (mm)
baseplate_pos = [0, 7.5, support_height];  // [X, Y, Z] offset from support (mm)

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
pi_hole_spacing = [
    2 * (pi_size[0]/2 - pi_hole_offsets[0]),  // 49mm (X/width)
    2 * (pi_size[1]/2 - pi_hole_offsets[1])   // 58mm (Y/length)
];

// Total assembly height
total_height = support_height + baseplate_thickness;  // 11mm

// ============================================
// BASEPLATE BODY
// ============================================
// Main platform for mounting Raspberry Pi

module baseplate_body() {
    // Platform is [width, length, thickness] = [X, Y, Z]
    safe_corner_radius = min(baseplate_corner_radius, baseplate_size[0] / 2, baseplate_size[1] / 2);

    linear_extrude(height=baseplate_thickness, center=true)
        offset(r=safe_corner_radius)
        offset(delta=-safe_corner_radius)
            square([baseplate_size[0], baseplate_size[1]], center=true);
}

// ============================================
// RASPBERRY PI MOUNTING HOLES
// ============================================
// Four M2.5 screw holes matching Pi 3B+ pattern

module pi_mounting_holes() {
    // Calculate hole positions from baseplate center
    hole_pos = [
        baseplate_size[0]/2 - pi_hole_offsets[0],  // 24.5mm from center
        baseplate_size[1]/2 - pi_hole_offsets[1]   // 29mm from center
    ];
    
    for (x = [-1, 1]) {
        for (y = [-1, 1]) {
            translate([
                x * hole_pos[0],
                y * hole_pos[1],
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
            translate([baseplate_pos[0], baseplate_pos[1], baseplate_pos[2] + baseplate_thickness/2])
                baseplate_body();
            
            // Orientation marker (small notch on front edge)
            translate([baseplate_pos[0], baseplate_pos[1] + baseplate_size[1]/2 - 2, baseplate_pos[2] + baseplate_thickness - 0.5])
                rotate([0, 45, 0])
                cube([1, 3, 1], center=true);
        }
        
        // Cut mounting holes from baseplate
        translate([baseplate_pos[0], baseplate_pos[1], baseplate_pos[2] + baseplate_thickness/2])
            pi_mounting_holes();
    }
}

// ============================================
// MAIN RENDER
// ============================================
baseplate_initial();