select
	trim(name || ' ' || version) as name,
	sum(count_unique)            as count_unique
from system_stats
join systems using (system_id)
where
	site_id = :site and day >= :start and day <= :end and
	{{:filter path_id in (:filter) and}}
	lower(name) = lower(:system)
group by name, version
order by count_unique desc, name asc
limit :limit offset :offset
