local allowed_types = {
    MUNICIPAL = true,
    COOPERATIVE = true,
    ["INVESTOR OWNED"] = true
}

function filterna(v)
    if v == "NOT AVAILABLE" then
        return nil
    end
    return v
end

function title(s)
    -- TODO: Implement ME!
    return s
end

function process(feature)
    if allowed_types[feature.properties.OWNER_TYPE] then
        emit({
            id = pluscode(centroid(feature.geometry), 6),
            geometry = simplify(feature.geometry, 0.001),
            properties = {
                source = "https://opendata.arcgis.com/datasets/c4fd0b01c2544a2f83440dab292f0980_0.geojson",
                objectid = "utilities." .. tostring(feature.properties.OBJECTID),
                area_sqft = sqm2sqft(area(feature.geometry)),
                name = feature.properties.NAME,
                owner_type = title(feature.properties.OWNER_TYPE),
                holding_company = title(feature.properties.HOLDING_CO),
                url = filterna(feature.properties.WEBSITE),
                total_customers = feature.properties.CUSTOMERS,
                control_area = title(feature.properties.CNTRL_AREA),
                plan_area = title(feature.properties.PLAN_AREA),
                regulated = title(feature.properties.REGULATED),
                summer_peak_load_mw = feature.properties.SUMMR_PEAK,
                winter_peak_load_mw = feature.properties.WINTR_PEAK,
                summer_capacity_mw = feature.properties.SUMMER_CAP,
                winter_capacity_mw = feature.properties.WINTER_CAP,
                net_generation_mwh = feature.properties.NET_GEN,
                purchased_mwh = feature.properties.PURCHASED,
                net_export_mwh = feature.properties.NET_EX,
                retail_mwh = feature.properties.RETAIL_MWH,
                wholesale_mwh = feature.properties.WSALE_MWH,
                total_mwh = feature.properties.TOTAL_MWH,
                address = {
                    street_address = title(feature.properties.ADDRESS),
                    locality = title(feature.properties.CITY),
                    region = feature.properties.STATE,
                    postal_code = feature.properties.ZIPCODE,
                    country = feature.properties.COUNTRY
                },
                telephone = feature.properties.PHONE,
                year = feature.properties.YEAR,
                source = feature.properties.SOURCE
            }
        })
    end
end
