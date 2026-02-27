/*
 * Nomad Core - X306 + Pi Case Brainplate
 * Standalone brainplate with recessed cutout pocket
 */

// ============================================
// IMPORTS
// ============================================
use <../parts/screw_cutout.scad>

// ============================================
// BRAINPLATE CONFIGURATION
// ============================================
// All brainplate settings consolidated here for easy adjustment:
// - Outer body dimensions
// - Pocket cutout dimensions and floor thickness
// - Wall naming/orientation
// - Interior feature dimensions and offsets

// ============================================
// MAIN DIMENSIONS
// ============================================
base_size = [58.75, 85.0, 4.8];     // [Width(X), Length(Y), Height(Z)] (mm)
cutout_size = [53.25, 79.5, 40.0];  // [Width(X), Length(Y), Height(Z)] (mm)

// ============================================
// CUTOUT PLACEMENT
// ============================================
cutout_floor_z = 1.0;       // Leaves a 1.0 mm floor thickness in base block
cutout_border_x_left = 3.5; // Left long-wall border from outer edge (mm)

// ============================================
// WALL NAMING CONVENTION (TOP VIEW)
// ============================================
// - Left wall: long wall near X=0
// - Right wall: thin wall near X=max
// - Bottom wall: short wall near Y=0
// - Top wall: short wall near Y=max

// ============================================
// FEATURE DIMENSIONS
// ============================================
right_feature_size = [15.0, 55.0, 2.0];  // [Depth(X), Length(Y), Height(Z)] (mm)
right_feature_offset_from_top = 4.0;     // Offset from inner top wall (mm)

left_feature_size = [29.5, cutout_size[1], right_feature_size[2]];  // [Depth(X), Length(Y), Height(Z)] (mm)

// ============================================
// SCREW HOLE CONFIGURATION
// ============================================
// Uses shared screw_cutout() standard module
screw_offset_from_left_top = [4.0, 4.0];  // [Inner left wall, Inner top wall] center offsets (mm)
screw_pi_vertical_spacing = 58.0;        // Standard Pi hole spacing, top-to-bottom (mm)
screw_pi_horizontal_spacing = 23;     // Standard Pi hole spacing, left-to-right (mm)
screw_inset_from_bottom = -12;           // Vertical inset to cut out screw hole without the nut recess (mm)

// ============================================
// RENDER QUALITY
// ============================================
$fn = 64;

// ============================================
// CALCULATED VALUES
// ============================================
// Cutout origin from outer body
cutout_pos = [
    cutout_border_x_left,
    (base_size[1] - cutout_size[1]) / 2,
    cutout_floor_z
];

// Outer walls from top-view orientation
left_wall_x = 0;
right_wall_x = base_size[0];
bottom_wall_y = 0;
top_wall_y = base_size[1];

// Inner pocket wall coordinates
inner_left_x = cutout_pos[0];
inner_right_x = cutout_pos[0] + cutout_size[0];
inner_bottom_y = cutout_pos[1];
inner_top_y = cutout_pos[1] + cutout_size[1];

// Right-wall feature placement
// - Flush to inner right wall and pocket floor
// - Offset from inner top wall
right_feature_pos = [
    inner_right_x - right_feature_size[0],
    inner_top_y - right_feature_offset_from_top - right_feature_size[1],
    cutout_floor_z
];

// Left-wall feature placement
// - Flush to inner left wall
// - Spans full inner length from bottom to top
// - Same floor and height as right-wall feature
left_feature_pos = [
    inner_left_x,
    inner_bottom_y,
    cutout_floor_z
];

// Screw hole placement (from inner pocket walls)
screw_hole_pos = [
    inner_left_x + screw_offset_from_left_top[0],
    inner_top_y - screw_offset_from_left_top[1],
    screw_inset_from_bottom
];

// Second screw hole placement (straight down from first)
screw_hole_2_pos = [
    screw_hole_pos[0],
    screw_hole_pos[1] - screw_pi_vertical_spacing,
    screw_hole_pos[2]
];

// Third screw hole placement (22.75mm right of first)
screw_hole_3_pos = [
    screw_hole_pos[0] + screw_pi_horizontal_spacing,
    screw_hole_pos[1],
    screw_hole_pos[2]
];

// Fourth screw hole placement (22.75mm right of second)
screw_hole_4_pos = [
    screw_hole_2_pos[0] + screw_pi_horizontal_spacing,
    screw_hole_2_pos[1],
    screw_hole_2_pos[2]
];

// ============================================
// MODEL
// ============================================
module x306_pi_brainplate() {
    difference() {
        union() {
            difference() {
                cube(base_size, center=false);
                translate(cutout_pos)
                    cube(cutout_size, center=false);
            }

            translate(right_feature_pos)
                cube(right_feature_size, center=false);

            translate(left_feature_pos)
                cube(left_feature_size, center=false);
        }

        translate(screw_hole_pos)
            screw_cutout();

        translate(screw_hole_2_pos)
            screw_cutout();

        translate(screw_hole_3_pos)
            screw_cutout();

        translate(screw_hole_4_pos)
            screw_cutout();
    }
}

// ============================================
// MAIN RENDER
// ============================================
x306_pi_brainplate();
