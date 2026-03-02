/*
 * Nomad Core - Stream Deck Interface Plate v0.1
 * Initial interface plate scaffold for Stream Deck MK2 integration
 *
 * This model starts with a rounded plate and supports subtracting a
 * Stream Deck silhouette cutout from a converted reference STL.
 */

// ============================================
// IMPORTS
// ============================================
use <../module_mount_slim.scad>

// Convert `.reference/streamdeck.mk2.stp` to `.reference/streamdeck.mk2.stl`
// before enabling `use_reference_cutout`.

// ============================================
// INTERFACE PLATE CONFIGURATION
// ============================================
// All interface plate settings consolidated here for easy adjustment:
// - Plate dimensions and corner radius
// - Stream Deck cutout controls (position, scale, clearance)
// - Optional mounting holes

// ============================================
// MAIN DIMENSIONS
// ============================================
plate_size = [85.0, 120.0];          // [Width, Length] (mm)
plate_thickness = 25.0;              // Plate thickness / height (mm)
plate_corner_radius = 6.0;           // Outer corner radius (mm)
plate_rot = [0, 0, 0];             // [X, Y, Z] rotation (degrees, top-down clockwise)

// ============================================
// STREAM DECK CUTOUT SETTINGS
// ============================================
use_reference_cutout = true;         // Apply boolean cutout to plate geometry
show_cutout_preview = false;          // Show/hide Stream Deck reference preview
reference_stl_path = "../../.reference/streamdeck.mk2.stl";

cutout_pos = [0, 0, 27];           // [X, Y, Z] relative to plate center (mm)
cutout_rot = [0, 0, 90];           // [X, Y, Z] rotation (degrees, top-down clockwise)
cutout_scale = [1.0, 1.0, 1.0];      // [X, Y, Z] scale factor
use_silhouette_cutout = false;       // false = subtract real 3D model, true = subtract silhouette extrusion
cutout_clearance = 0.35;             // Additional XY clearance around cutout (mm)
cutout_apply_z = -1;               // Z used for boolean cut through the plate (mm)

// ============================================
// WEDGE CUTOUT 1 SETTINGS
// ============================================
use_wedge_cutout_1 = true;           // Apply wedge cutout 1 to plate geometry
show_wedge_cutout_1_preview = true;  // Show/hide wedge cutout 1 preview
wedge_cutout_1_pos = [63.5, 0, -1.0];   // [X, Y, Z] placement (mm)
wedge_cutout_1_rot = [0, 0, 0];      // [X, Y, Z] rotation (degrees)
wedge_cutout_1_width = 125.0;        // Width of +X-facing side (mm)
wedge_cutout_1_length = 127.0;       // Nose length toward -X (mm)
wedge_cutout_1_height = 14.0;        // Height (mm)

// ============================================
// MOUNTPLATE WING GHOST PREVIEW
// ============================================
show_mountplate_preview = true;      // Show/hide mountplate wing preview
mountplate_wing_pos = [-60, 35, 0];     // [X, Y, Z] preview position offset (mm)
mountplate_wing_rot = [-15, 180, -90];     // [X, Y, Z] preview rotation (degrees)
mountplate_wing_scale = [1, 1, 1];   // [X, Y, Z] preview scale

// ============================================
// OPTIONAL MOUNTING HOLES
// ============================================
add_mounting_holes = false;          // Toggle corner mounting holes
mount_hole_dia = 3.5;                // Hole diameter (mm)
mount_hole_inset = [10.0, 10.0];     // [X, Y] inset from plate edges (mm)

// ============================================
// RENDER QUALITY
// ============================================
preview_fn = 20;   // Faster viewport interaction
export_fn = 48;    // Better quality for STL export
$fn = $preview ? preview_fn : export_fn;

// ============================================
// HELPER MODULES
// ============================================

module rounded_plate_2d(size, radius) {
    safe_radius = min(radius, size[0] / 2, size[1] / 2);

    offset(r=safe_radius)
        offset(delta=-safe_radius)
            square(size, center=true);
}

module interface_plate_body() {
    linear_extrude(height=plate_thickness, center=false)
        rounded_plate_2d(plate_size, plate_corner_radius);
}

module streamdeck_cutout_volume() {
    if (use_silhouette_cutout) {
        translate([cutout_pos[0], cutout_pos[1], cutout_apply_z])
            rotate(cutout_rot)
                scale(cutout_scale)
                    linear_extrude(height=plate_thickness + 1.0, center=false)
                        offset(delta=cutout_clearance)
                            projection(cut=false)
                                import(reference_stl_path, convexity=10);
    } else {
        translate(cutout_pos)
            rotate(cutout_rot)
                scale(cutout_scale)
                    import(reference_stl_path, convexity=10);
    }
}

module streamdeck_reference_preview() {
    translate(cutout_pos)
        rotate(cutout_rot)
            scale(cutout_scale)
                import(reference_stl_path, convexity=10);
}

module mountplate_wing_reference_preview() {
    module_mount();
}

module wedge_cutout_1_volume() {
    width = wedge_cutout_1_width;
    length = wedge_cutout_1_length;
    height = wedge_cutout_1_height;

    translate(wedge_cutout_1_pos)
        rotate(wedge_cutout_1_rot)
            polyhedron(
                points=[
                    [0, -width / 2, 0],
                    [0, width / 2, 0],
                    [0, width / 2, height],
                    [0, -width / 2, height],
                    [-length, -width / 2, 0],
                    [-length, width / 2, 0]
                ],
                faces=[
                    [0, 1, 2, 3],
                    [0, 4, 5, 1],
                    [0, 3, 4],
                    [1, 5, 2],
                    [3, 2, 5, 4]
                ]
            );
}

module mounting_holes() {
    hole_x = plate_size[0] / 2 - mount_hole_inset[0];
    hole_y = plate_size[1] / 2 - mount_hole_inset[1];

    for (x = [-1, 1]) {
        for (y = [-1, 1]) {
            translate([x * hole_x, y * hole_y, -0.1])
                cylinder(h=plate_thickness + 0.4, d=mount_hole_dia);
        }
    }
}

// ============================================
// COMPLETE INTERFACE PLATE
// ============================================
module streamdeck_interface_plate() {
    difference() {
        interface_plate_body();

        if (use_reference_cutout) {
            streamdeck_cutout_volume();
        }

        if (use_wedge_cutout_1) {
            wedge_cutout_1_volume();
        }

        if (add_mounting_holes) {
            mounting_holes();
        }
    }
}

// ============================================
// MAIN RENDER
// ============================================
rotate(plate_rot)
    streamdeck_interface_plate();

if ($preview && show_cutout_preview) {
    #color([0.2, 0.7, 1.0, 0.6])
        rotate(plate_rot)
            streamdeck_reference_preview();
}

if ($preview && show_mountplate_preview) {
    #color([0.8, 0.8, 0.8, 0.6])
        translate(mountplate_wing_pos)
            rotate(mountplate_wing_rot)
                scale(mountplate_wing_scale)
                    mountplate_wing_reference_preview();
}

if ($preview && show_wedge_cutout_1_preview) {
    #color([1.0, 0.8, 0.2, 0.5])
        rotate(plate_rot)
            wedge_cutout_1_volume();
}
