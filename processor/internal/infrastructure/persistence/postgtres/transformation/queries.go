package transformation

const (
	insertTransformation = `
INSERT INTO transformations (
	id,
	image_id,

	source_storage_key,
	source_mime_type,
	source_width,
	source_height,

	transform_spec,
	transform_hash,

	status,

	result_storage_key,
	result_mime_type,
	result_width,
	result_height,
	result_size,

	error_message,

	started_at,
	completed_at,

	created_at,
	updated_at
)
VALUES (
	$1, $2,
	$3, $4, $5, $6,
	$7, $8,
	$9,
	$10, $11, $12, $13, $14,
	$15,
	$16, $17,
	$18, $19
);
`

	updateTransformation = `
UPDATE transformations
SET
	status = $2,

	result_storage_key = $3,
	result_mime_type = $4,
	result_width = $5,
	result_height = $6,
	result_size = $7,

	error_message = $8,

	started_at = $9,
	completed_at = $10,

	updated_at = $11
WHERE id = $1;
`

	getTransformationByID = `
SELECT
	id,
	image_id,

	source_storage_key,
	source_mime_type,
	source_width,
	source_height,

	transform_spec,
	transform_hash,

	status,

	result_storage_key,
	result_mime_type,
	result_width,
	result_height,
	result_size,

	error_message,

	started_at,
	completed_at,

	created_at,
	updated_at
FROM transformations
WHERE id = $1;
`

	getTransformationByImageAndHash = `
SELECT
	id,
	image_id,

	source_storage_key,
	source_mime_type,
	source_width,
	source_height,

	transform_spec,
	transform_hash,

	status,

	result_storage_key,
	result_mime_type,
	result_width,
	result_height,
	result_size,

	error_message,

	started_at,
	completed_at,

	created_at,
	updated_at
FROM transformations
WHERE image_id = $1
AND transform_hash = $2;
`

	acquireNextPending = `
WITH next_job AS (
	SELECT id
	FROM transformations
	WHERE status = 'pending'
	ORDER BY created_at
	LIMIT 1
	FOR UPDATE SKIP LOCKED
)
SELECT
	t.id,
	t.image_id,

	t.source_storage_key,
	t.source_mime_type,
	t.source_width,
	t.source_height,

	t.transform_spec,
	t.transform_hash,

	t.status,

	t.result_storage_key,
	t.result_mime_type,
	t.result_width,
	t.result_height,
	t.result_size,

	t.error_message,

	t.started_at,
	t.completed_at,

	t.created_at,
	t.updated_at
FROM transformations t
JOIN next_job j
	ON j.id = t.id;
`
)
