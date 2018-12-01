-- The examples deomonstrates the use of pluscode(point, len) to generate a pluscode [1] from a
-- single point (in this case, the centroid of the input geometry). The value of `len` determines
-- the length (and thus, precision) of the output pluscode.
--
-- [1] https://plus.codes

function process(feature)
    emit({
        id = pluscode(centroid(feature.geometry), 8),
        geometry = simplify(feature.geometry, 0.001),
        properties = feature.properties
    })
end
