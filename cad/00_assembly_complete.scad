/*
 * Nomad Core - Complete Assembly v1.0
 * Full system assembly view with all components
 * 
 * This file imports and positions all components for visualization
 * Supports both assembled and exploded views with configurable positioning
 */

// ============================================
// IMPORTS
// ============================================
use <puck_design.scad>
use <baseplate_initial.scad>
use <module_mount_slim.scad>

// ============================================
// ASSEMBLY CONFIGURATION
// ============================================
// All assembly settings are consolidated here for easy adjustment:
// - Component visibility toggles (show_*)
// - View mode (assembled vs exploded)
// - Component positions ([X, Y, Z] arrays)
// - Component rotations ([X, Y, Z] arrays in degrees)
// - Merge pieces (bridge geometry between components)

// ============================================
// COMPONENT VISIBILITY TOGGLES
// ============================================
show_puck = true;          // Toggle puck visibility
show_baseplate = true;     // Toggle baseplate visibility
show_module_mount = true;   // Toggle module mount visibility

// ============================================
// VIEW CONFIGURATION
// ============================================
show_exploded = false;     // Set to true for exploded view
explode_distance = 20;     // Separation distance for exploded view (mm)

// ============================================
// COMPONENT POSITIONS
// ============================================
// Puck position (base reference point)
puck_pos = [0, 0, 0];  // [X, Y, Z] (mm)

// Baseplate position (sits on top of puck)
baseplate_pos = [0, 0, 3.5];  // [X, Y, Z] - Z=3.5 is puck height (mm)

// Module mount position (left side of baseplate)
module_mount_left_pos = [35, 42.5, 14.9];  // [X, Y, Z] (mm)

// Opposite module mount position (right side of baseplate)
module_mount_right_pos = [-35, -27.5, 14.9];  // [X, Y, Z] (mm)

// ============================================
// COMPONENT ROTATIONS
// ============================================
// Puck rotation
puck_rot = [0, 0, 0];  // [X, Y, Z] rotation in degrees

// Baseplate rotation
baseplate_rot = [0, 0, 0];  // [X, Y, Z] rotation in degrees

// Module mount rotation (60° slant for ergonomics)
module_mount_rot = [60, 180, -90];  // [X, Y, Z] rotation in degrees

// Opposite module mount rotation (mirrored side)
module_mount_opposite_rot = [60, 180, 90];  // [X, Y, Z] rotation in degrees

// ============================================
// MERGE PIECES CONFIGURATION
// ============================================
// Bridge components together with custom geometry
// Format: [shape, [width, length, height], [x, y, z], [rot_x, rot_y, rot_z]]
// Available shapes: "cuboid", "wedge"

merge_pieces = [
    // Left module mount to baseplate bridge
    [
        "cuboid",
        [5.5, 70, 4],           // Dimensions [W, L, H] (mm)
        [29.1, -27.5, 10.5],   // Position [X, Y, Z] (mm)
        [0, 0, 0]             // Rotation [X, Y, Z] (degrees)
    ],
    // Right module mount to baseplate bridge
    [
        "cuboid",
        [5.5, 70, 4],            // Dimensions [W, L, H] (mm)
        [-34.6, -27.5, 10.5],    // Position [X, Y, Z] (mm)
        [0, 0, 0]                // Rotation [X, Y, Z] (degrees)
    ]
];

// ============================================
// RENDER QUALITY
// ============================================
$fn = 64;

// ============================================
// MERGE PIECE HELPER MODULES
// ============================================

// Simple cuboid for bridging components
module merge_cuboid(width, length, height) {
    cube([width, length, height], center=false);
}

// Simple wedge (right-angle triangular prism)
module merge_wedge(width, length, height) {
    linear_extrude(height=length)
        polygon([
            [0, 0],
            [width, 0],
            [0, height]
        ]);
}

// Render a merge piece from array parameters
module render_merge_piece(shape, dims, pos, rot) {
    translate(pos)
        rotate(rot) {
            if (shape == "cuboid") {
                merge_cuboid(dims[0], dims[1], dims[2]);
            } else if (shape == "wedge") {
                merge_wedge(dims[0], dims[1], dims[2]);
            }
        }
}

// ============================================
// MAIN ASSEMBLY
// ============================================

// Puck (base component)
if (show_puck) {
    translate(puck_pos)
        rotate(puck_rot)
            color("SteelBlue")
                puck_design();
}

// Baseplate (sits on top of puck)
if (show_baseplate) {
    baseplate_offset = show_exploded ? explode_distance : 0;
    
    translate([baseplate_pos[0], baseplate_pos[1], baseplate_pos[2] + baseplate_offset])
        rotate(baseplate_rot)
            color("Tomato")
                baseplate_initial();
}

// Module mounts (both sides of baseplate)
if (show_module_mount) {
    module_mount_offset = show_exploded ? explode_distance : 0;
    
    // Left module mount
    translate([module_mount_left_pos[0], module_mount_left_pos[1], module_mount_left_pos[2] + module_mount_offset])
        rotate(module_mount_rot)
            color("LimeGreen")
                module_mount();

    // Right module mount (opposite side)
    translate([
        module_mount_right_pos[0],
        module_mount_right_pos[1],
        module_mount_right_pos[2] + module_mount_offset
    ])
        rotate(module_mount_opposite_rot)
            color("LimeGreen")
                module_mount();
}

// Merge pieces (bridge components together)
if (len(merge_pieces) > 0) {
    for (i = [0 : len(merge_pieces) - 1]) {
        piece = merge_pieces[i];
        color("Orange")
            render_merge_piece(piece[0], piece[1], piece[2], piece[3]);
    }
}