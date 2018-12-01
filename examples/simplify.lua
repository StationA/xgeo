-- The examples deomonstrates the use of simplify(geom, threshold) to simplify the input geometry
-- using the well-known Douglas-Peucker algorithm [1].
--
-- [1] https://en.wikipedia.org/wiki/Ramer%E2%80%93Douglas%E2%80%93Peucker_algorithm

function process(feature)
    emit({
        id = feature.id,
        geometry = simplify(feature.geometry, 0.001),
        properties = feature.properties
    })
end
