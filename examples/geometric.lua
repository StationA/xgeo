function process(feature)
    emit({
        id = feature.id,
        geometry = feature.geometry,
        properties = {
            area_sqm = area(feature.geometry),
            area_sqft = sqm2sqft(area(feature.geometry)),
            perimeter_m = perimeter(feature.geometry),
            perimeter_ft = m2ft(perimeter(feature.geometry))
        }
    })
end
