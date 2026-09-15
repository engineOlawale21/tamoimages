package database

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Database struct{ pool *pgxpool.Pool }

type MediaAsset struct {
	ID               string
	ContributorID    string
	Kind             string
	OriginalFilename string
	StorageKey       string
	Status           string
	ContentType      string
	SizeBytes        int64
	DurationSeconds  float64
	Width            int
	Height           int
	CodecName        string
	FailureCode      string
	Title            string
	Description      string
	Keywords         []string
	Location         string
	UsageType        string
	ProcessedAt      *time.Time
	CreatedAt        time.Time
	UpdatedAt        time.Time
}

type ProcessingMetadata struct {
	DurationSeconds float64
	Width           int
	Height          int
	CodecName       string
}

type MediaVariant struct {
	Kind        string
	StorageKey  string
	ContentType string
	SizeBytes   int64
	Width       int
	Height      int
}

type MediaBatch struct {
	ID             string
	ContributorID  string
	Name           string
	Status         string
	ItemCount      int
	SubmittedAt    *time.Time
	ReviewFeedback string
	ReviewedAt     *time.Time
	CreatedAt      time.Time
	UpdatedAt      time.Time
}

type MediaRelease struct {
	ID, MediaAssetID, ContributorID, ReleaseType, Filename, StorageKey, ContentType, Status string
	SizeBytes                                                                               int64
	UploadExpiresAt                                                                         time.Time
	CreatedAt, UpdatedAt                                                                    time.Time
}
type MediaDeletionTargets struct {
	AssetID, ContributorID string
	StorageKeys            []string
}
type BuyerCollection struct {
	ID, BuyerID, Name string
	MediaIDs          []string
}

type CatalogQuery struct {
	Text, Kind, UsageType, Orientation, Location, Sort string
	Page, PageSize                                     int
}

func scanMediaAsset(row pgx.Row, asset *MediaAsset) error {
	return row.Scan(&asset.ID, &asset.ContributorID, &asset.Kind, &asset.OriginalFilename, &asset.StorageKey,
		&asset.Status, &asset.ContentType, &asset.SizeBytes, &asset.DurationSeconds, &asset.Width,
		&asset.Height, &asset.CodecName, &asset.FailureCode, &asset.Title, &asset.Description, &asset.Keywords,
		&asset.Location, &asset.UsageType, &asset.ProcessedAt, &asset.CreatedAt, &asset.UpdatedAt)
}

const mediaAssetColumns = `a.id,a.contributor_id,a.kind,a.original_filename,a.storage_key,a.status,
	a.content_type,a.size_bytes,COALESCE(a.duration_seconds,0),COALESCE(a.width,0),COALESCE(a.height,0),
	COALESCE(a.codec_name,''),COALESCE(a.failure_code,''),COALESCE(a.title,''),COALESCE(a.description,''),a.keywords,
	COALESCE(a.location,''),COALESCE(a.usage_type,''),a.processed_at,a.created_at,a.updated_at`

func (d *Database) CatalogAssets(ctx context.Context, query CatalogQuery) ([]MediaAsset, int, error) {
	const predicates = `a.status='ready' AND a.deleted_at IS NULL AND b.status='approved' AND b.deleted_at IS NULL
		AND ($1='' OR to_tsvector('simple',coalesce(a.title,'')||' '||coalesce(a.description,'')||' '||coalesce(a.location,'')) @@ plainto_tsquery('simple',$1) OR lower($1)=ANY(a.keywords))
		AND ($2='' OR a.kind=$2) AND ($3='' OR a.usage_type=$3)
		AND ($4='' OR ($4='landscape' AND a.width>a.height) OR ($4='portrait' AND a.height>a.width) OR ($4='square' AND a.width=a.height))
		AND ($5='' OR a.location ILIKE '%'||$5||'%')`
	var total int
	err := d.pool.QueryRow(ctx, `SELECT count(DISTINCT a.id) FROM media_assets a JOIN media_batch_items i ON i.media_asset_id=a.id JOIN media_batches b ON b.id=i.batch_id WHERE `+predicates,
		query.Text, query.Kind, query.UsageType, query.Orientation, query.Location).Scan(&total)
	if err != nil {
		return nil, 0, fmt.Errorf("count catalog assets: %w", err)
	}
	order := "a.created_at DESC, a.id DESC"
	if query.Sort == "relevance" && query.Text != "" {
		order = "ts_rank(to_tsvector('simple',coalesce(a.title,'')||' '||coalesce(a.description,'')||' '||coalesce(a.location,'')),plainto_tsquery('simple',$1)) DESC,a.created_at DESC"
	}
	rows, err := d.pool.Query(ctx, `SELECT `+mediaAssetColumns+` FROM media_assets a JOIN media_batch_items i ON i.media_asset_id=a.id JOIN media_batches b ON b.id=i.batch_id WHERE `+predicates+` ORDER BY `+order+` LIMIT $6 OFFSET $7`,
		query.Text, query.Kind, query.UsageType, query.Orientation, query.Location, query.PageSize, (query.Page-1)*query.PageSize)
	if err != nil {
		return nil, 0, fmt.Errorf("search catalog assets: %w", err)
	}
	defer rows.Close()
	assets := make([]MediaAsset, 0)
	for rows.Next() {
		var asset MediaAsset
		if err = scanMediaAsset(rows, &asset); err != nil {
			return nil, 0, fmt.Errorf("scan catalog asset: %w", err)
		}
		assets = append(assets, asset)
	}
	if err = rows.Err(); err != nil {
		return nil, 0, fmt.Errorf("read catalog assets: %w", err)
	}
	return assets, total, nil
}

func (d *Database) CatalogAssetByID(ctx context.Context, id string) (MediaAsset, error) {
	var asset MediaAsset
	err := scanMediaAsset(d.pool.QueryRow(ctx, `SELECT `+mediaAssetColumns+` FROM media_assets a JOIN media_batch_items i ON i.media_asset_id=a.id JOIN media_batches b ON b.id=i.batch_id WHERE a.id=$1 AND a.status='ready' AND a.deleted_at IS NULL AND b.status='approved' AND b.deleted_at IS NULL`, id), &asset)
	if err != nil {
		return MediaAsset{}, fmt.Errorf("get catalog asset: %w", err)
	}
	return asset, nil
}

func Open(ctx context.Context, connectionString string, maximumConnections int32, connectTimeout time.Duration) (*Database, error) {
	config, err := pgxpool.ParseConfig(connectionString)
	if err != nil {
		return nil, fmt.Errorf("parse database configuration: %w", err)
	}
	config.MaxConns = maximumConnections
	config.MinConns = 0
	config.MaxConnIdleTime = 5 * time.Minute
	config.MaxConnLifetime = 30 * time.Minute
	connectContext, cancel := context.WithTimeout(ctx, connectTimeout)
	defer cancel()
	pool, err := pgxpool.NewWithConfig(connectContext, config)
	if err != nil {
		return nil, fmt.Errorf("create database pool: %w", err)
	}
	return &Database{pool: pool}, nil
}

func (d *Database) Close()                         { d.pool.Close() }
func (d *Database) Ping(ctx context.Context) error { return d.pool.Ping(ctx) }

func (d *Database) CreateMediaAsset(ctx context.Context, asset MediaAsset) (MediaAsset, error) {
	err := d.pool.QueryRow(ctx, `INSERT INTO media_assets
		(id,contributor_id,kind,original_filename,storage_key,status,content_type,size_bytes)
		VALUES($1,$2,$3,$4,$5,'pending',$6,$7)
		RETURNING status,created_at,updated_at`, asset.ID, asset.ContributorID, asset.Kind,
		asset.OriginalFilename, asset.StorageKey, asset.ContentType, asset.SizeBytes,
	).Scan(&asset.Status, &asset.CreatedAt, &asset.UpdatedAt)
	if err != nil {
		return MediaAsset{}, fmt.Errorf("create media asset: %w", err)
	}
	return asset, nil
}

func (d *Database) MediaAssetByID(ctx context.Context, id, contributorID string) (MediaAsset, error) {
	var asset MediaAsset
	err := d.pool.QueryRow(ctx, `SELECT id,contributor_id,kind,original_filename,storage_key,status,
		content_type,size_bytes,COALESCE(duration_seconds,0),COALESCE(width,0),COALESCE(height,0),
		COALESCE(codec_name,''),COALESCE(failure_code,''),COALESCE(title,''),COALESCE(description,''),keywords,
		COALESCE(location,''),COALESCE(usage_type,''),processed_at,created_at,updated_at FROM media_assets
		WHERE id=$1 AND contributor_id=$2 AND deleted_at IS NULL`, id, contributorID).Scan(
		&asset.ID, &asset.ContributorID, &asset.Kind, &asset.OriginalFilename, &asset.StorageKey,
		&asset.Status, &asset.ContentType, &asset.SizeBytes, &asset.DurationSeconds, &asset.Width,
		&asset.Height, &asset.CodecName, &asset.FailureCode, &asset.Title, &asset.Description, &asset.Keywords,
		&asset.Location, &asset.UsageType, &asset.ProcessedAt, &asset.CreatedAt, &asset.UpdatedAt,
	)
	if err != nil {
		return MediaAsset{}, fmt.Errorf("get media asset: %w", err)
	}
	return asset, nil
}

func (d *Database) MediaAssetsByContributor(ctx context.Context, contributorID string, limit int) ([]MediaAsset, error) {
	rows, err := d.pool.Query(ctx, `SELECT id,contributor_id,kind,original_filename,storage_key,status,
		content_type,size_bytes,COALESCE(duration_seconds,0),COALESCE(width,0),COALESCE(height,0),
		COALESCE(codec_name,''),COALESCE(failure_code,''),COALESCE(title,''),COALESCE(description,''),keywords,
		COALESCE(location,''),COALESCE(usage_type,''),processed_at,created_at,updated_at FROM media_assets
		WHERE contributor_id=$1 AND deleted_at IS NULL ORDER BY created_at DESC LIMIT $2`, contributorID, limit)
	if err != nil {
		return nil, fmt.Errorf("list media assets: %w", err)
	}
	defer rows.Close()
	assets := make([]MediaAsset, 0)
	for rows.Next() {
		var asset MediaAsset
		if err = rows.Scan(&asset.ID, &asset.ContributorID, &asset.Kind, &asset.OriginalFilename, &asset.StorageKey,
			&asset.Status, &asset.ContentType, &asset.SizeBytes, &asset.DurationSeconds, &asset.Width,
			&asset.Height, &asset.CodecName, &asset.FailureCode, &asset.Title, &asset.Description, &asset.Keywords,
			&asset.Location, &asset.UsageType, &asset.ProcessedAt, &asset.CreatedAt, &asset.UpdatedAt); err != nil {
			return nil, fmt.Errorf("scan media asset: %w", err)
		}
		assets = append(assets, asset)
	}
	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("read media assets: %w", err)
	}
	return assets, nil
}

func (d *Database) MediaVariantsByAssetID(ctx context.Context, id string) ([]MediaVariant, error) {
	rows, err := d.pool.Query(ctx, `SELECT kind,storage_key,content_type,size_bytes,COALESCE(width,0),COALESCE(height,0)
		FROM media_variants WHERE media_asset_id=$1 ORDER BY
		CASE kind WHEN 'thumbnail' THEN 1 WHEN 'poster' THEN 2 WHEN 'small' THEN 3 WHEN 'medium' THEN 4 WHEN 'large' THEN 5 ELSE 6 END`, id)
	if err != nil {
		return nil, fmt.Errorf("get media variants: %w", err)
	}
	defer rows.Close()
	variants := make([]MediaVariant, 0)
	for rows.Next() {
		var variant MediaVariant
		if err = rows.Scan(&variant.Kind, &variant.StorageKey, &variant.ContentType, &variant.SizeBytes, &variant.Width, &variant.Height); err != nil {
			return nil, fmt.Errorf("scan media variant: %w", err)
		}
		variants = append(variants, variant)
	}
	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("read media variants: %w", err)
	}
	return variants, nil
}

func (d *Database) MediaDeletionTargets(ctx context.Context, id, contributorID string) (MediaDeletionTargets, error) {
	var target MediaDeletionTargets
	var original string
	err := d.pool.QueryRow(ctx, `SELECT id,contributor_id,storage_key FROM media_assets WHERE id=$1 AND contributor_id=$2 AND deleted_at IS NULL`, id, contributorID).Scan(&target.AssetID, &target.ContributorID, &original)
	if err != nil {
		return MediaDeletionTargets{}, err
	}
	target.StorageKeys = append(target.StorageKeys, original)
	rows, err := d.pool.Query(ctx, `SELECT storage_key FROM media_variants WHERE media_asset_id=$1 UNION ALL SELECT storage_key FROM media_releases WHERE media_asset_id=$1`, id)
	if err != nil {
		return MediaDeletionTargets{}, fmt.Errorf("list deletion targets: %w", err)
	}
	defer rows.Close()
	for rows.Next() {
		var key string
		if err = rows.Scan(&key); err != nil {
			return MediaDeletionTargets{}, err
		}
		target.StorageKeys = append(target.StorageKeys, key)
	}
	return target, rows.Err()
}

func (d *Database) CompleteMediaDeletion(ctx context.Context, target MediaDeletionTargets, reason, targetHash string) error {
	tx, err := d.pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)
	result, err := tx.Exec(ctx, `UPDATE media_assets SET status='deleted',deleted_at=now(),updated_at=now(),original_filename='deleted',storage_key='deleted/'||id::text,content_type=NULL,size_bytes=NULL,title=NULL,description=NULL,keywords='{}',location=NULL,codec_name=NULL,failure_code=NULL WHERE id=$1 AND contributor_id=$2 AND deleted_at IS NULL`, target.AssetID, target.ContributorID)
	if err != nil {
		return fmt.Errorf("mark media deleted: %w", err)
	}
	if result.RowsAffected() != 1 {
		return pgx.ErrNoRows
	}
	if _, err = tx.Exec(ctx, `DELETE FROM media_batch_items WHERE media_asset_id=$1; DELETE FROM media_variants WHERE media_asset_id=$1; DELETE FROM media_releases WHERE media_asset_id=$1`, target.AssetID); err != nil {
		return fmt.Errorf("delete media relations: %w", err)
	}
	if _, err = tx.Exec(ctx, `INSERT INTO media_deletion_evidence(media_asset_id,contributor_id,reason,object_count,target_hash) VALUES($1,$2,$3,$4,$5)`, target.AssetID, target.ContributorID, reason, len(target.StorageKeys), targetHash); err != nil {
		return fmt.Errorf("record deletion evidence: %w", err)
	}
	return tx.Commit(ctx)
}

func (d *Database) CreateBuyerCollection(ctx context.Context, id, buyerID, name string) (BuyerCollection, error) {
	var result BuyerCollection
	err := d.pool.QueryRow(ctx, `INSERT INTO buyer_collections(id,buyer_id,name) VALUES($1,$2,$3) RETURNING id,buyer_id,name`, id, buyerID, name).Scan(&result.ID, &result.BuyerID, &result.Name)
	return result, err
}
func (d *Database) BuyerCollections(ctx context.Context, buyerID string) ([]BuyerCollection, error) {
	rows, err := d.pool.Query(ctx, `SELECT c.id,c.buyer_id,c.name,COALESCE(array_agg(i.media_asset_id::text ORDER BY i.added_at) FILTER(WHERE i.media_asset_id IS NOT NULL),'{}') FROM buyer_collections c LEFT JOIN buyer_collection_items i ON i.collection_id=c.id WHERE c.buyer_id=$1 GROUP BY c.id ORDER BY c.created_at DESC`, buyerID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	result := []BuyerCollection{}
	for rows.Next() {
		var item BuyerCollection
		if err = rows.Scan(&item.ID, &item.BuyerID, &item.Name, &item.MediaIDs); err != nil {
			return nil, err
		}
		result = append(result, item)
	}
	return result, rows.Err()
}
func (d *Database) AddBuyerCollectionItem(ctx context.Context, buyerID, collectionID, assetID string) (bool, error) {
	result, err := d.pool.Exec(ctx, `INSERT INTO buyer_collection_items(collection_id,media_asset_id) SELECT c.id,a.id FROM buyer_collections c CROSS JOIN media_assets a WHERE c.id=$1 AND c.buyer_id=$2 AND a.id=$3 AND a.status='ready' AND a.deleted_at IS NULL AND EXISTS(SELECT 1 FROM media_batch_items i JOIN media_batches b ON b.id=i.batch_id WHERE i.media_asset_id=a.id AND b.status='approved' AND b.deleted_at IS NULL) ON CONFLICT DO NOTHING`, collectionID, buyerID, assetID)
	return result.RowsAffected() > 0, err
}
func (d *Database) RemoveBuyerCollectionItem(ctx context.Context, buyerID, collectionID, assetID string) (bool, error) {
	result, err := d.pool.Exec(ctx, `DELETE FROM buyer_collection_items i USING buyer_collections c WHERE i.collection_id=c.id AND c.id=$1 AND c.buyer_id=$2 AND i.media_asset_id=$3`, collectionID, buyerID, assetID)
	return result.RowsAffected() > 0, err
}
func (d *Database) BuyerFavourites(ctx context.Context, buyerID string) ([]string, error) {
	rows, err := d.pool.Query(ctx, `SELECT media_asset_id::text FROM buyer_favourites WHERE buyer_id=$1 ORDER BY added_at DESC`, buyerID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	result := []string{}
	for rows.Next() {
		var id string
		if err = rows.Scan(&id); err != nil {
			return nil, err
		}
		result = append(result, id)
	}
	return result, rows.Err()
}
func (d *Database) SetBuyerFavourite(ctx context.Context, buyerID, assetID string, value bool) (bool, error) {
	if !value {
		_, err := d.pool.Exec(ctx, `DELETE FROM buyer_favourites WHERE buyer_id=$1 AND media_asset_id=$2`, buyerID, assetID)
		return true, err
	}
	result, err := d.pool.Exec(ctx, `INSERT INTO buyer_favourites(buyer_id,media_asset_id) SELECT $1,a.id FROM media_assets a WHERE a.id=$2 AND a.status='ready' AND a.deleted_at IS NULL AND EXISTS(SELECT 1 FROM media_batch_items i JOIN media_batches b ON b.id=i.batch_id WHERE i.media_asset_id=a.id AND b.status='approved' AND b.deleted_at IS NULL) ON CONFLICT DO NOTHING`, buyerID, assetID)
	return result.RowsAffected() > 0, err
}

func (d *Database) UpdateMediaAssetMetadata(ctx context.Context, id, contributorID string, metadata MediaAsset) (MediaAsset, error) {
	result, err := d.pool.Exec(ctx, `UPDATE media_assets SET title=$3,description=$4,keywords=$5,location=$6,usage_type=$7,updated_at=now()
		WHERE id=$1 AND contributor_id=$2 AND status='ready' AND deleted_at IS NULL`, id, contributorID,
		metadata.Title, nullableString(metadata.Description), metadata.Keywords, nullableString(metadata.Location), metadata.UsageType)
	if err != nil {
		return MediaAsset{}, fmt.Errorf("update media metadata: %w", err)
	}
	if result.RowsAffected() != 1 {
		return MediaAsset{}, pgx.ErrNoRows
	}
	return d.MediaAssetByID(ctx, id, contributorID)
}

func (d *Database) CreateMediaBatch(ctx context.Context, batch MediaBatch) (MediaBatch, error) {
	err := d.pool.QueryRow(ctx, `INSERT INTO media_batches(id,contributor_id,name) VALUES($1,$2,$3)
		RETURNING status,submitted_at,created_at,updated_at`, batch.ID, batch.ContributorID, batch.Name).Scan(
		&batch.Status, &batch.SubmittedAt, &batch.CreatedAt, &batch.UpdatedAt)
	if err != nil {
		return MediaBatch{}, fmt.Errorf("create media batch: %w", err)
	}
	return batch, nil
}

func (d *Database) MediaBatchesByContributor(ctx context.Context, contributorID string, limit int) ([]MediaBatch, error) {
	rows, err := d.pool.Query(ctx, `SELECT b.id,b.contributor_id,b.name,b.status,count(i.media_asset_id),b.submitted_at,COALESCE(b.review_feedback,''),b.reviewed_at,b.created_at,b.updated_at
		FROM media_batches b LEFT JOIN media_batch_items i ON i.batch_id=b.id
		WHERE b.contributor_id=$1 AND b.deleted_at IS NULL GROUP BY b.id ORDER BY b.created_at DESC LIMIT $2`, contributorID, limit)
	if err != nil {
		return nil, fmt.Errorf("list media batches: %w", err)
	}
	defer rows.Close()
	batches := make([]MediaBatch, 0)
	for rows.Next() {
		var batch MediaBatch
		if err = rows.Scan(&batch.ID, &batch.ContributorID, &batch.Name, &batch.Status, &batch.ItemCount, &batch.SubmittedAt, &batch.ReviewFeedback, &batch.ReviewedAt, &batch.CreatedAt, &batch.UpdatedAt); err != nil {
			return nil, fmt.Errorf("scan media batch: %w", err)
		}
		batches = append(batches, batch)
	}
	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("read media batches: %w", err)
	}
	return batches, nil
}

func (d *Database) AddMediaAssetToBatch(ctx context.Context, batchID, contributorID, assetID string) (bool, error) {
	result, err := d.pool.Exec(ctx, `INSERT INTO media_batch_items(batch_id,media_asset_id)
		SELECT b.id,a.id FROM media_batches b JOIN media_assets a ON a.id=$3
		WHERE b.id=$1 AND b.contributor_id=$2 AND b.status='draft' AND b.deleted_at IS NULL
		AND a.contributor_id=$2 AND a.status='ready' AND a.deleted_at IS NULL
		ON CONFLICT (media_asset_id) DO NOTHING`, batchID, contributorID, assetID)
	if err != nil {
		return false, fmt.Errorf("add media asset to batch: %w", err)
	}
	if result.RowsAffected() == 1 {
		return true, nil
	}
	var assigned bool
	err = d.pool.QueryRow(ctx, `SELECT EXISTS(SELECT 1 FROM media_batch_items i JOIN media_batches b ON b.id=i.batch_id
		WHERE i.batch_id=$1 AND i.media_asset_id=$2 AND b.contributor_id=$3 AND b.status='draft')`, batchID, assetID, contributorID).Scan(&assigned)
	if err != nil {
		return false, fmt.Errorf("check media batch assignment: %w", err)
	}
	return assigned, nil
}

func (d *Database) SubmitMediaBatch(ctx context.Context, batchID, contributorID string) (MediaBatch, error) {
	var batch MediaBatch
	err := d.pool.QueryRow(ctx, `UPDATE media_batches b SET status='submitted',submitted_at=now(),updated_at=now()
		WHERE b.id=$1 AND b.contributor_id=$2 AND b.status='draft' AND b.deleted_at IS NULL
		AND EXISTS(SELECT 1 FROM media_batch_items i WHERE i.batch_id=b.id)
		AND NOT EXISTS(SELECT 1 FROM media_batch_items i JOIN media_assets a ON a.id=i.media_asset_id
			WHERE i.batch_id=b.id AND (a.title IS NULL OR a.usage_type IS NULL OR cardinality(a.keywords)=0))
		RETURNING b.id,b.contributor_id,b.name,b.status,(SELECT count(*) FROM media_batch_items i WHERE i.batch_id=b.id),b.submitted_at,COALESCE(b.review_feedback,''),b.reviewed_at,b.created_at,b.updated_at`, batchID, contributorID).Scan(
		&batch.ID, &batch.ContributorID, &batch.Name, &batch.Status, &batch.ItemCount, &batch.SubmittedAt, &batch.ReviewFeedback, &batch.ReviewedAt, &batch.CreatedAt, &batch.UpdatedAt)
	if err != nil {
		return MediaBatch{}, fmt.Errorf("submit media batch: %w", err)
	}
	return batch, nil
}

func (d *Database) MediaBatchByID(ctx context.Context, batchID string) (MediaBatch, error) {
	var batch MediaBatch
	err := d.pool.QueryRow(ctx, `SELECT b.id,b.contributor_id,b.name,b.status,count(i.media_asset_id),b.submitted_at,COALESCE(b.review_feedback,''),b.reviewed_at,b.created_at,b.updated_at
		FROM media_batches b LEFT JOIN media_batch_items i ON i.batch_id=b.id WHERE b.id=$1 AND b.deleted_at IS NULL GROUP BY b.id`, batchID).Scan(&batch.ID, &batch.ContributorID, &batch.Name, &batch.Status, &batch.ItemCount, &batch.SubmittedAt, &batch.ReviewFeedback, &batch.ReviewedAt, &batch.CreatedAt, &batch.UpdatedAt)
	if err != nil {
		return MediaBatch{}, fmt.Errorf("get media batch: %w", err)
	}
	return batch, nil
}

func (d *Database) ReviewMediaBatch(ctx context.Context, batchID, status, feedback string) (MediaBatch, error) {
	var batch MediaBatch
	err := d.pool.QueryRow(ctx, `UPDATE media_batches b SET status=$2,review_feedback=$3,reviewed_at=CASE WHEN $2 IN ('approved','rejected') THEN now() ELSE NULL END,updated_at=now()
		WHERE b.id=$1 AND b.status IN ('submitted','in_review') AND b.deleted_at IS NULL
		RETURNING b.id,b.contributor_id,b.name,b.status,(SELECT count(*) FROM media_batch_items i WHERE i.batch_id=b.id),b.submitted_at,COALESCE(b.review_feedback,''),b.reviewed_at,b.created_at,b.updated_at`, batchID, status, nullableString(feedback)).Scan(&batch.ID, &batch.ContributorID, &batch.Name, &batch.Status, &batch.ItemCount, &batch.SubmittedAt, &batch.ReviewFeedback, &batch.ReviewedAt, &batch.CreatedAt, &batch.UpdatedAt)
	if err != nil {
		return MediaBatch{}, fmt.Errorf("review media batch: %w", err)
	}
	return batch, nil
}

func (d *Database) CreateMediaRelease(ctx context.Context, release MediaRelease) (MediaRelease, bool, error) {
	tx, err := d.pool.Begin(ctx)
	if err != nil {
		return MediaRelease{}, false, fmt.Errorf("begin media release: %w", err)
	}
	defer func() { _ = tx.Rollback(ctx) }()
	var assetID string
	if err = tx.QueryRow(ctx, `SELECT id FROM media_assets WHERE id=$1 AND contributor_id=$2 AND status='ready' AND deleted_at IS NULL FOR UPDATE`, release.MediaAssetID, release.ContributorID).Scan(&assetID); err != nil {
		return MediaRelease{}, false, err
	}
	if _, err = tx.Exec(ctx, `DELETE FROM media_releases WHERE media_asset_id=$1 AND status='pending' AND upload_expires_at < now()`, release.MediaAssetID); err != nil {
		return MediaRelease{}, false, fmt.Errorf("remove expired media releases: %w", err)
	}
	var count int
	if err = tx.QueryRow(ctx, `SELECT count(*) FROM media_releases WHERE media_asset_id=$1`, release.MediaAssetID).Scan(&count); err != nil {
		return MediaRelease{}, false, fmt.Errorf("count media releases: %w", err)
	}
	if count >= 3 {
		return MediaRelease{}, true, nil
	}
	err = tx.QueryRow(ctx, `INSERT INTO media_releases(id,media_asset_id,contributor_id,release_type,filename,storage_key,content_type,size_bytes,upload_expires_at)
		VALUES($1,$2,$3,$4,$5,$6,$7,$8,$9) RETURNING status,created_at,updated_at`, release.ID, release.MediaAssetID, release.ContributorID,
		release.ReleaseType, release.Filename, release.StorageKey, release.ContentType, release.SizeBytes, release.UploadExpiresAt).Scan(&release.Status, &release.CreatedAt, &release.UpdatedAt)
	if err != nil {
		return MediaRelease{}, false, fmt.Errorf("create media release: %w", err)
	}
	if err = tx.Commit(ctx); err != nil {
		return MediaRelease{}, false, fmt.Errorf("commit media release: %w", err)
	}
	return release, false, nil
}

func (d *Database) MediaReleaseByID(ctx context.Context, assetID, releaseID, contributorID string) (MediaRelease, error) {
	var release MediaRelease
	err := d.pool.QueryRow(ctx, `SELECT id,media_asset_id,contributor_id,release_type,filename,storage_key,content_type,size_bytes,status,upload_expires_at,created_at,updated_at
		FROM media_releases WHERE id=$1 AND media_asset_id=$2 AND contributor_id=$3`, releaseID, assetID, contributorID).Scan(&release.ID, &release.MediaAssetID, &release.ContributorID, &release.ReleaseType, &release.Filename, &release.StorageKey, &release.ContentType, &release.SizeBytes, &release.Status, &release.UploadExpiresAt, &release.CreatedAt, &release.UpdatedAt)
	if err != nil {
		return MediaRelease{}, fmt.Errorf("get media release: %w", err)
	}
	return release, nil
}

func (d *Database) MarkMediaReleaseUploaded(ctx context.Context, assetID, releaseID, contributorID string) (MediaRelease, error) {
	var release MediaRelease
	err := d.pool.QueryRow(ctx, `UPDATE media_releases SET status='uploaded',updated_at=now() WHERE id=$1 AND media_asset_id=$2 AND contributor_id=$3 AND status='pending' AND upload_expires_at>=now()
		RETURNING id,media_asset_id,contributor_id,release_type,filename,storage_key,content_type,size_bytes,status,upload_expires_at,created_at,updated_at`, releaseID, assetID, contributorID).Scan(&release.ID, &release.MediaAssetID, &release.ContributorID, &release.ReleaseType, &release.Filename, &release.StorageKey, &release.ContentType, &release.SizeBytes, &release.Status, &release.UploadExpiresAt, &release.CreatedAt, &release.UpdatedAt)
	if err != nil {
		return MediaRelease{}, fmt.Errorf("complete media release: %w", err)
	}
	return release, nil
}

func (d *Database) MediaReleasesByAsset(ctx context.Context, assetID, contributorID string) ([]MediaRelease, error) {
	rows, err := d.pool.Query(ctx, `SELECT r.id,r.media_asset_id,r.contributor_id,r.release_type,r.filename,r.storage_key,r.content_type,r.size_bytes,r.status,r.upload_expires_at,r.created_at,r.updated_at
		FROM media_releases r JOIN media_assets a ON a.id=r.media_asset_id
		WHERE r.media_asset_id=$1 AND r.contributor_id=$2 AND a.contributor_id=$2 AND a.deleted_at IS NULL
		AND (r.status='uploaded' OR r.upload_expires_at>=now()) ORDER BY r.created_at`, assetID, contributorID)
	if err != nil {
		return nil, fmt.Errorf("list media releases: %w", err)
	}
	defer rows.Close()
	releases := make([]MediaRelease, 0)
	for rows.Next() {
		var release MediaRelease
		if err = rows.Scan(&release.ID, &release.MediaAssetID, &release.ContributorID, &release.ReleaseType, &release.Filename, &release.StorageKey, &release.ContentType, &release.SizeBytes, &release.Status, &release.UploadExpiresAt, &release.CreatedAt, &release.UpdatedAt); err != nil {
			return nil, fmt.Errorf("scan media release: %w", err)
		}
		releases = append(releases, release)
	}
	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("read media releases: %w", err)
	}
	return releases, nil
}

func (d *Database) MarkMediaAssetUploaded(ctx context.Context, id, contributorID string) (MediaAsset, error) {
	var asset MediaAsset
	err := d.pool.QueryRow(ctx, `UPDATE media_assets SET status='uploaded',updated_at=now()
		WHERE id=$1 AND contributor_id=$2 AND status='pending' AND deleted_at IS NULL
		RETURNING id,contributor_id,kind,original_filename,storage_key,status,
		content_type,size_bytes,created_at,updated_at`, id, contributorID).Scan(
		&asset.ID, &asset.ContributorID, &asset.Kind, &asset.OriginalFilename, &asset.StorageKey,
		&asset.Status, &asset.ContentType, &asset.SizeBytes, &asset.CreatedAt, &asset.UpdatedAt,
	)
	if err != nil {
		return MediaAsset{}, fmt.Errorf("mark media asset uploaded: %w", err)
	}
	return asset, nil
}

func (d *Database) RetryFailedMediaAsset(ctx context.Context, id, contributorID string) (MediaAsset, error) {
	var asset MediaAsset
	err := d.pool.QueryRow(ctx, `UPDATE media_assets SET status='uploaded',failure_code=NULL,processed_at=NULL,updated_at=now()
		WHERE id=$1 AND contributor_id=$2 AND status='failed' AND deleted_at IS NULL
		RETURNING id,contributor_id,kind,original_filename,storage_key,status,
		content_type,size_bytes,created_at,updated_at`, id, contributorID).Scan(
		&asset.ID, &asset.ContributorID, &asset.Kind, &asset.OriginalFilename, &asset.StorageKey,
		&asset.Status, &asset.ContentType, &asset.SizeBytes, &asset.CreatedAt, &asset.UpdatedAt,
	)
	if err != nil {
		return MediaAsset{}, fmt.Errorf("retry failed media asset: %w", err)
	}
	return asset, nil
}

func (d *Database) ClaimMediaAssetForProcessing(ctx context.Context, id string) (MediaAsset, bool, error) {
	var asset MediaAsset
	err := d.pool.QueryRow(ctx, `UPDATE media_assets SET status='processing',failure_code=NULL,updated_at=now()
		WHERE id=$1 AND deleted_at IS NULL AND
		(status='uploaded' OR (status='processing' AND updated_at < now() - interval '5 minutes'))
		RETURNING id,contributor_id,kind,original_filename,storage_key,status,
		content_type,size_bytes,created_at,updated_at`, id).Scan(
		&asset.ID, &asset.ContributorID, &asset.Kind, &asset.OriginalFilename, &asset.StorageKey,
		&asset.Status, &asset.ContentType, &asset.SizeBytes, &asset.CreatedAt, &asset.UpdatedAt,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return MediaAsset{}, false, nil
	}
	if err != nil {
		return MediaAsset{}, false, fmt.Errorf("claim media asset: %w", err)
	}
	return asset, true, nil
}

func (d *Database) CompleteMediaAssetProcessing(ctx context.Context, id string, metadata ProcessingMetadata, variants []MediaVariant) error {
	return d.InTransaction(ctx, func(tx pgx.Tx) error {
		for _, variant := range variants {
			if _, err := tx.Exec(ctx, `INSERT INTO media_variants(media_asset_id,kind,storage_key,content_type,size_bytes,width,height)
				VALUES($1,$2,$3,$4,$5,$6,$7)
				ON CONFLICT(media_asset_id,kind) DO UPDATE SET storage_key=excluded.storage_key,content_type=excluded.content_type,
				size_bytes=excluded.size_bytes,width=excluded.width,height=excluded.height`, id, variant.Kind, variant.StorageKey,
				variant.ContentType, variant.SizeBytes, nullableInt(variant.Width), nullableInt(variant.Height)); err != nil {
				return fmt.Errorf("record media variant: %w", err)
			}
		}
		result, err := tx.Exec(ctx, `UPDATE media_assets SET status='ready',duration_seconds=$2,width=$3,height=$4,
			codec_name=$5,failure_code=NULL,processed_at=now(),updated_at=now()
			WHERE id=$1 AND status='processing'`, id, nullableFloat(metadata.DurationSeconds), nullableInt(metadata.Width), nullableInt(metadata.Height), nullableString(metadata.CodecName))
		if err != nil {
			return fmt.Errorf("mark media asset ready: %w", err)
		}
		if result.RowsAffected() != 1 {
			return fmt.Errorf("mark media asset ready: asset is not processing")
		}
		return nil
	})
}

func (d *Database) MarkMediaAssetFailed(ctx context.Context, id, failureCode string) error {
	_, err := d.pool.Exec(ctx, `UPDATE media_assets SET status='failed',failure_code=$2,processed_at=now(),updated_at=now()
		WHERE id=$1 AND status='processing'`, id, failureCode)
	if err != nil {
		return fmt.Errorf("mark media asset failed: %w", err)
	}
	return nil
}

func nullableString(value string) any {
	if value == "" {
		return nil
	}
	return value
}
func nullableInt(value int) any {
	if value == 0 {
		return nil
	}
	return value
}
func nullableFloat(value float64) any {
	if value == 0 {
		return nil
	}
	return value
}

func (d *Database) InTransaction(ctx context.Context, operation func(pgx.Tx) error) error {
	transaction, err := d.pool.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return fmt.Errorf("begin transaction: %w", err)
	}
	defer func() { _ = transaction.Rollback(ctx) }()
	if err = operation(transaction); err != nil {
		return err
	}
	if err = transaction.Commit(ctx); err != nil {
		return fmt.Errorf("commit transaction: %w", err)
	}
	return nil
}

func (d *Database) Migrate(ctx context.Context, directory string) error {
	if _, err := d.pool.Exec(ctx, `CREATE TABLE IF NOT EXISTS schema_migrations (name text PRIMARY KEY, applied_at timestamptz NOT NULL DEFAULT now())`); err != nil {
		return fmt.Errorf("create migration metadata: %w", err)
	}
	entries, err := os.ReadDir(directory)
	if err != nil {
		return fmt.Errorf("read migrations: %w", err)
	}
	files := make([]string, 0, len(entries))
	for _, entry := range entries {
		if !entry.IsDir() && strings.HasSuffix(entry.Name(), ".up.sql") {
			files = append(files, entry.Name())
		}
	}
	sort.Strings(files)
	for _, name := range files {
		var applied bool
		if err = d.pool.QueryRow(ctx, `SELECT EXISTS (SELECT 1 FROM schema_migrations WHERE name=$1)`, name).Scan(&applied); err != nil {
			return err
		}
		if applied {
			continue
		}
		script, readErr := os.ReadFile(filepath.Join(directory, name))
		if readErr != nil {
			return readErr
		}
		if err = d.InTransaction(ctx, func(tx pgx.Tx) error {
			if _, executeErr := tx.Exec(ctx, string(script)); executeErr != nil {
				return executeErr
			}
			_, executeErr := tx.Exec(ctx, `INSERT INTO schema_migrations(name) VALUES($1)`, name)
			return executeErr
		}); err != nil {
			return fmt.Errorf("apply migration %s: %w", name, err)
		}
	}
	return nil
}
