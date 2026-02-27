/*
 * Nomad Core - Support Column Module
 * Reusable support structure with rounded edges
 * 
 * Parameters are passed as function arguments
 */

// ============================================
// SUPPORT COLUMN MODULE
// ============================================
// Parameters:
// - size: footprint size (e.g., 25mm)
// - height: column height (e.g., 7mm)
// - corner_radius: rounded corner radius (e.g., 3mm)

module support_column(size=25, height=7, corner_radius=3) {
    // Solid hull with rounded corners
    hull() {
        // Four rounded corners
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

support_column();