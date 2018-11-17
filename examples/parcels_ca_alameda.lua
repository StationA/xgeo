function process(feature)
    emit({
        id = pluscode(centroid(feature.geometry), 9),
        geometry = feature.geometry,
        properties = {
            source = "",
            objectid = "ca.alameda." .. feature.properties.apn,
            apn = feature.properties.apn,
            area_sqft = sqm2sqft(area(feature.geometry)),
            url = "http://www.acgov.org/MS/prop/index.aspx?PRINT_PARCEL=" .. feature.properties.apn
        }
    })
end
