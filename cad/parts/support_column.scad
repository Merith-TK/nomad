/*
 * Nomad Core - Support Column Module
 * Reusable support structure with rounded edges
 * 
 * Creates a solid rectangular column with rounded corners using hull operation
 */

// ============================================
// SUPPORT COLUMN CONFIGURATION
// ============================================
// Default parameters for test renders:
// - Footprint size (square)
// - Column height
// - Corner radius for rounding

// ============================================
// DEFAULT PARAMETERS
// ============================================
default_size = 25.0;           // Default footprint size (mm)
default_height = 7.0;          // Default column height (mm)
default_corner_radius = 3.0;   // Default corner radius (mm)

// ============================================
// RENDER QUALITY
// ============================================
$fn = 64;

// ============================================
// SUPPORT COLUMN MODULE
// ============================================
// Creates a solid support column with rounded corners
//
// Parameters:
// - size: Square footprint dimension (mm)
// - height: Column height (mm)
// - corner_radius: Radius of rounded corners (mm)
//
// Example: support_column(25, 7, 3);

module support_column(size=25, height=7, corner_radius=3) {
    // Solid hull with rounded corners
    hull() {
        // Four rounded corners positioned at calculated offsets
        for (x = [-1, 1]) {
            for (y = [-1, 1]) {
                translate([
                    x * (size/2 - corner_radius),
                    y * (size/2 - corner_radius),
                    0
                ])
                    cylinder(h=height, r=corner_radius, center=true);
            }
        }
    }
}

// ============================================
// TEST RENDER (uncomment for visualization)
// ============================================
// support_column(default_size, default_height, default_corner_radius);