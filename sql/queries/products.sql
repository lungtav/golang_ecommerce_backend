-- name: CreateProduct :one
insert into products (store_id, name, status, description) values ($1,$2,$3,$4) returning *;

-- name: CreateProductVariant :one
insert into product_variants (product_id, sku, size, color, price_in_kobo) values ($1, $2, $3, $4, $5) returning *;

-- name: GetProduct :one
select p.id, p.store_id, p.name, p.status, p.description, p.created_at,
    coalesce(
        jsonb_agg(
            jsonb_build_object(
                'sku', pv.sku,
                'size', pv.size,
                'color', pv.color,
                'price_in_kobo', pv.price_in_kobo,
                'inventory', i.quantity
            )
        ) filter (where pv.id is not null),
        '[]'::jsonb
    ) as variants from products as p
left join product_variants as pv
    on p.id = pv.product_id left join inventories as i
    on i.product_variant_id = pv.id
    and i.store_id = p.store_id
    where p.id = $1
group by p.id, p.store_id, p.name, p.status, p.description, p.created_at;