/*
 * Nomad Core - X306 + Pi Case Brainplate
 * Standalone brainplate with recessed cutout pocket
 */

// ============================================
// BRAINPLATE CONFIGURATION
// ============================================
// Outer block dimensions (adjusted to satisfy perimeter constraints)
base_width = 58.75;     // X dimension (mm) -> gives 3.5 mm + 2.0 mm side borders around 53.25 mm cutout
base_length = 85.0;     // Y dimension (mm)
base_height = 4.8;      // Z dimension (mm)

// Cutout block dimensions
cutout_width = 53.25;   // X dimension (mm)
cutout_length = 79.5;   // Y dimension (mm)
cutout_height = 40.0;   // Tall enough to fully pass through above cut floor (mm)

// Cutout placement
cutout_floor_z = 1.0;   // Leaves a 1.0 mm floor thickness in base block
cutout_border_x_left = 3.5; // Border on one long side wall (mm)

// Derived placement
cutout_x = cutout_border_x_left;                         // Leaves 2.0 mm on opposite long side wall
cutout_y = (base_length - cutout_length) / 2;           // Symmetric short-end borders (>= 2 mm)

// Inner pocket wall coordinates
inner_left_x = cutout_x;
inner_right_x = cutout_x + cutout_width;
inner_bottom_y = cutout_y;
inner_top_y = cutout_y + cutout_length;

// ============================================
// WALL NAMING CONVENTION
// ============================================
// Top-view naming:
// - Left wall: long wall near X=0
// - Right wall: thin wall near X=max
// - Bottom wall: short wall near Y=0
// - Top wall: short wall near Y=max

left_wall_x = 0;
right_wall_x = base_width;
bottom_wall_y = 0;
top_wall_y = base_length;

// ============================================
// RIGHT-WALL FEATURE
// ============================================
// Rectangular feature in the pocket:
// - 3 mm from top wall
// - 15 mm depth from right wall
// - 58 mm long
// - 2 mm tall
// - starts at Z=1 so it sits on the pocket floor

right_feature_offset_from_top = 4.0;
right_feature_depth = 15.0;
right_feature_length = 55.0;
right_feature_height = 2.0;
right_feature_z = cutout_floor_z;

// Flush to inner pocket right wall and floor, offset 4 mm from inner top wall
right_feature_x = inner_right_x - right_feature_depth;
right_feature_y = inner_top_y - right_feature_offset_from_top - right_feature_length;

// ============================================
// LEFT-WALL FEATURE
// ============================================
// Rectangular feature in the pocket:
// - flush to inner left wall
// - spans full inner length
// - 29.5 mm wide (extends inward from left wall)
// - same height and floor alignment as right-wall feature

left_feature_width = 29.5;
left_feature_length = cutout_length;
left_feature_height = right_feature_height;
left_feature_z = cutout_floor_z;

left_feature_x = inner_left_x;
left_feature_y = inner_bottom_y;

// ============================================
// RENDER QUALITY
// ============================================
$fn = 64;

// ============================================
// MODEL
// ============================================
module x306_pi_brainplate() {
    union() {
        difference() {
            cube([base_width, base_length, base_height], center=false);
            translate([cutout_x, cutout_y, cutout_floor_z])
                cube([cutout_width, cutout_length, cutout_height], center=false);
        }

        translate([right_feature_x, right_feature_y, right_feature_z])
            cube([right_feature_depth, right_feature_length, right_feature_height], center=false);

        translate([left_feature_x, left_feature_y, left_feature_z])
            cube([left_feature_width, left_feature_length, left_feature_height], center=false);
    }
}

// ============================================
// MAIN RENDER
// ============================================
x306_pi_brainplate();
