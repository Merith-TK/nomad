/*
 * Nomad Core - Complete Assembly v1.0
 * All-in-one view of the full system
 * 
 * This file imports and positions all components
 */

// Import component modules (use = modules only, don't render)
use <puck_design.scad>
use <baseplate_initial.scad>
use <brain_mount.scad>

// ============================================
// ASSEMBLY CONFIGURATION
// ============================================
show_puck = true;
show_baseplate = true;
show_brain_mount = true;
show_exploded = false;  // Set to true for exploded view

explode_distance = 20;  // Distance for exploded view

// ============================================
// PUCK POSITION & ROTATION
// ============================================
puck_pos_x = 0;
puck_pos_y = 0;
puck_pos_z = 0;
puck_rot_x = 0;
puck_rot_y = 0;
puck_rot_z = 0;

// ============================================
// BASEPLATE POSITION & ROTATION
// ============================================
baseplate_pos_x = 0;
baseplate_pos_y = 0;
baseplate_pos_z = 3.5;  // Puck height (default: sits on top of puck)
baseplate_rot_x = 0;
baseplate_rot_y = 0;
baseplate_rot_z = 0;

// ============================================
// BRAIN MOUNT POSITION & ROTATION
// ============================================
brain_mount_pos_x = 34;  // Left side of baseplate
brain_mount_pos_y = 42.5;
brain_mount_pos_z = 14.5;  // Same height as baseplate (default)
brain_mount_rot_x = 60;
brain_mount_rot_y = 180;
brain_mount_rot_z = -90;

// ============================================
// MERGE PIECES (Bridge components together)
// ============================================
// Array format: [shape, [width, length, height], [x, y, z], [rot_x, rot_y, rot_z]]
// shape: "cuboid" or "wedge"
// Example:
// ["cuboid", [10, 20, 5], [0, 0, 3.5], [0, 0, 0]]
// ["wedge", [15, 25, 7], [-30, 0, 3.5], [0, 0, 90]]

merge_pieces = [
    // Add merge pieces here as needed
    // [
    //   "type", 
    //   [dimx, dimy, dimz],
    //   [posx, posy, posz],
    //   [rotx, roty, rotz]
    // ]
    // ["cuboid", [10, 20, 5], [0, 0, 3.5], [0, 0, 0]]
    [
        "cuboid", // Join Brain Plate with Base Plate
        [5, 70, 5], 
        [29.1, -27.5, 9.5], 
        [0, 0, 0]
    ]
];

// ============================================
// RENDER QUALITY
// ============================================
$fn = 64;

// ============================================
// HELPER MODULES FOR MERGE PIECES
// ============================================

// Simple cuboid
module merge_cuboid(width, length, height) {
    cube([width, length, height], center=false);
}

// Simple wedge (right-angle triangular prism)
module merge_wedge(width, length, height) {
    // width = base of triangle
    // length = extrusion depth
    // height = height of triangle
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
// COMPONENT POSITIONING
// ============================================

// Puck (base component)
if (show_puck) {
    translate([puck_pos_x, puck_pos_y, puck_pos_z])
        rotate([puck_rot_x, puck_rot_y, puck_rot_z])
            color("SteelBlue")
                puck_design();
}

// Baseplate (sits on top of puck by default)
if (show_baseplate) {
    baseplate_offset = show_exploded ? explode_distance : 0;
    
    translate([baseplate_pos_x, baseplate_pos_y, baseplate_pos_z + baseplate_offset])
        rotate([baseplate_rot_x, baseplate_rot_y, baseplate_rot_z])
            color("Tomato")
                baseplate_initial();
}

// Brain Mount (left side of baseplate by default)
if (show_brain_mount) {
    brain_mount_offset = show_exploded ? explode_distance : 0;
    
    translate([brain_mount_pos_x, brain_mount_pos_y, brain_mount_pos_z + brain_mount_offset])
        rotate([brain_mount_rot_x, brain_mount_rot_y, brain_mount_rot_z])
            color("LimeGreen")
                brain_mount();
}

// Merge Pieces (bridge components together)
if (len(merge_pieces) > 0) {
    for (i = [0 : len(merge_pieces) - 1]) {
        piece = merge_pieces[i];
        color("Orange")
            render_merge_piece(piece[0], piece[1], piece[2], piece[3]);
    }
}

// ============================================
// INFO OUTPUT
// ============================================
echo("=== NOMAD CORE ASSEMBLY v1.0 ===");
echo(str("Components: ", (show_puck ? "puck " : ""), (show_baseplate ? "baseplate " : ""), (show_brain_mount ? "brain_mount " : "")));
echo(str("Merge pieces: ", len(merge_pieces)));
echo(str("View mode: ", show_exploded ? "EXPLODED" : "ASSEMBLED"));
if (show_exploded) {
    echo(str("Explode distance: ", explode_distance, " mm"));
}
echo("");
echo("=== COMPONENT POSITIONS ===");
if (show_puck) {
    echo(str("Puck: [", puck_pos_x, ", ", puck_pos_y, ", ", puck_pos_z, "] | Rot: [", puck_rot_x, ", ", puck_rot_y, ", ", puck_rot_z, "]"));
}
if (show_baseplate) {
    echo(str("Baseplate: [", baseplate_pos_x, ", ", baseplate_pos_y, ", ", baseplate_pos_z, "] | Rot: [", baseplate_rot_x, ", ", baseplate_rot_y, ", ", baseplate_rot_z, "]"));
}
if (show_brain_mount) {
    echo(str("Brain Mount: [", brain_mount_pos_x, ", ", brain_mount_pos_y, ", ", brain_mount_pos_z, "] | Rot: [", brain_mount_rot_x, ", ", brain_mount_rot_y, ", ", brain_mount_rot_z, "]"));
}
echo("");
echo("TIP: Set show_exploded = true for exploded view");
echo("TIP: Toggle components with show_* variables");
echo("TIP: Adjust position with *_pos_x/y/z variables");
echo("TIP: Adjust rotation with *_rot_x/y/z variables");
echo("TIP: Add merge pieces: [\"cuboid\"|\"wedge\", [w,l,h], [x,y,z], [rx,ry,rz]]");
