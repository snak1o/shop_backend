package repository

import (
	"context"
	"database/sql"
	"fmt"
	"github.com/jmoiron/sqlx"
	"shop_backend/internal/models"
)

type ItemsRepo struct {
	db *sqlx.DB
}

func NewItemsRepo(db *sqlx.DB) *ItemsRepo {
	return &ItemsRepo{db: db}
}

func (r *ItemsRepo) WithinTransaction(ctx context.Context, tFunc func(ctx context.Context) error) error {
	tx, err := r.db.Beginx()
	if err != nil {
		return fmt.Errorf("begin transcation: %w", err)
	}

	if err := tFunc(injectTx(ctx, tx)); err != nil {
		tx.Rollback()
		return err
	}
	tx.Commit()
	return nil
}

func (r *ItemsRepo) GetInstance(ctx context.Context) SqlxDB {
	tx := extractTx(ctx)
	if tx != nil {
		return tx
	}
	return r.db
}

func (r *ItemsRepo) Create(ctx context.Context, item models.Item) (int, error) {
	db := r.GetInstance(ctx)
	var id int
	query := fmt.Sprintf("INSERT INTO %s (name,description,category_id,sku,price) VALUES ($1,$2,$3,$4,$5) RETURNING id;", itemsTable)
	if err := db.GetContext(ctx, &id, query, item.Name, item.Description, item.Category.Id, item.Sku, item.Price); err != nil {
		return 0, err
	}

	return id, nil
}

func (r *ItemsRepo) LinkColor(ctx context.Context, itemId int, colorId int) error {
	db := r.GetInstance(ctx)
	query := fmt.Sprintf("INSERT INTO %s (item_id,color_id) VALUES ($1,$2);", itemsColorsTable)
	_, err := db.ExecContext(ctx, query, itemId, colorId)

	return err
}

func (r *ItemsRepo) LinkTag(ctx context.Context, itemId int, tag string) error {
	db := r.GetInstance(ctx)
	query := fmt.Sprintf("INSERT INTO %s (item_id, name) VALUES($1,$2);", tagsTable)
	_, err := db.ExecContext(ctx, query, itemId, tag)

	return err
}

func (r *ItemsRepo) LinkImage(ctx context.Context, itemId, imageId int) error {
	db := r.GetInstance(ctx)
	query := fmt.Sprintf("INSERT INTO %s (item_id, image_id) VALUES ($1, $2);", itemsImagesTable)
	_, err := db.ExecContext(ctx, query, itemId, imageId)

	return err
}

func (r *ItemsRepo) GetNew(ctx context.Context, limit int) ([]int, error) {
	var ids []int
	query := fmt.Sprintf("SELECT I.id FROM %s AS I ORDER BY created_at DESC LIMIT $1;", itemsTable)
	if err := r.db.Select(&ids, query, limit); err != nil {
		return nil, err
	}

	return ids, nil
}

func (r *ItemsRepo) GetAll(ctx context.Context, sortOptions models.SortOptions) ([]int, error) {
	var ids []int
	query := fmt.Sprintf("SELECT I.id FROM %s AS I ORDER BY %s %s;", itemsTable, sortOptions.Field, sortOptions.Order)
	if err := r.db.Select(&ids, query); err != nil {
		return nil, err
	}

	return ids, nil
}

func (r *ItemsRepo) GetById(ctx context.Context, itemId int) (models.Item, error) {
	db := r.GetInstance(ctx)
	var item models.Item
	query := fmt.Sprintf("SELECT * FROM %s WHERE id=$1;", itemsTable)
	if err := db.QueryRowContext(ctx, query, itemId).Scan(&item.Id, &item.Name, &item.Description, &item.Category.Id, &item.Price, &item.Sku, &item.CreatedAt); err != nil {
		return models.Item{}, err
	}

	return item, nil
}

func (r *ItemsRepo) GetBySku(ctx context.Context, sku string) (models.Item, error) {
	var item models.Item
	query := fmt.Sprintf("SELECT * FROM %s where sku=$1;", itemsTable)
	if err := r.db.QueryRow(query, sku).Scan(&item.Id, &item.Name, &item.Description, &item.Category.Id, &item.Price, &item.Sku, &item.CreatedAt); err != nil {
		return models.Item{}, err
	}

	return item, nil
}

func (r *ItemsRepo) GetByCategory(ctx context.Context, categoryId int) ([]int, error) {
	var ids []int
	query := fmt.Sprintf("SELECT I.id FROM %s AS I WHERE category_id=$1;", itemsTable)
	if err := r.db.Select(&ids, query, categoryId); err != nil {
		return nil, err
	}

	return ids, nil
}

func (r *ItemsRepo) GetByTag(ctx context.Context, tag string) ([]int, error) {
	var ids []int
	query := fmt.Sprintf("SELECT I.id FROM %s AS I, %s AS T WHERE T.name = $1 AND I.id = T.item_id;", itemsTable, tagsTable)
	if err := r.db.Select(&ids, query, tag); err != nil {
		return nil, err
	}

	return ids, nil
}

func (r *ItemsRepo) GetColors(ctx context.Context, itemId int) ([]models.Color, error) {
	db := r.GetInstance(ctx)
	var colors []models.Color
	query := fmt.Sprintf("SELECT colors.id, colors.name, colors.hex, colors.price FROM %s, %s WHERE colors.id = %s.color_id AND %s.item_id = $1;", colorsTable, itemsColorsTable, itemsColorsTable, itemsColorsTable)
	rows, err := db.QueryxContext(ctx, query, itemId)
	if err != nil {
		return nil, err
	}

	for rows.Next() {
		var c models.Color
		if err := rows.StructScan(&c); err != nil {
			return nil, err
		}
		colors = append(colors, c)
	}

	return colors, nil
}

func (r *ItemsRepo) GetTags(ctx context.Context, itemId int) ([]models.Tag, error) {
	db := r.GetInstance(ctx)
	var tags []models.Tag
	query := fmt.Sprintf("SELECT * FROM %s WHERE tags.item_id = $1;", tagsTable)
	rows, err := db.QueryxContext(ctx, query, itemId)
	if err != nil {
		return nil, err
	}

	for rows.Next() {
		var t models.Tag
		if err := rows.StructScan(&t); err != nil {
			return nil, err
		}
		tags = append(tags, t)
	}

	return tags, nil
}

func (r *ItemsRepo) GetImages(ctx context.Context, itemId int) ([]models.Image, error) {
	db := r.GetInstance(ctx)
	var images []models.Image
	query := fmt.Sprintf("SELECT images.id, images.filename, images.created_at FROM %s, %s WHERE images.id = %s.image_id AND %s.item_id = $1;", imagesTable, itemsImagesTable, itemsImagesTable, itemsImagesTable)
	rows, err := db.QueryxContext(ctx, query, itemId)
	if err != nil {
		return nil, err
	}

	for rows.Next() {
		var i models.Image
		if err := rows.StructScan(&i); err != nil {
			return nil, err
		}
		images = append(images, i)
	}

	return images, nil
}

func (r *ItemsRepo) Update(ctx context.Context, itemId int, name, description string, categoryId int, price float64, sku string) error {
	query := fmt.Sprintf("UPDATE %s SET name=$1,description=$2,category_id=$3,price=$4,sku=$5 WHERE id=$6;", itemsTable)
	_, err := r.db.Exec(query, name, description, categoryId, price, sku, itemId)

	return err
}

func (r *ItemsRepo) Delete(ctx context.Context, itemId int) error {
	db := r.GetInstance(ctx)
	query := fmt.Sprintf("DELETE FROM %s WHERE id=$1", itemsTable)
	_, err := db.ExecContext(ctx, query, itemId)

	return err
}

func (r *ItemsRepo) DeleteTags(ctx context.Context, itemId int) error {
	query := fmt.Sprintf("DELETE FROM %s WHERE item_id=$1;", tagsTable)
	_, err := r.db.Exec(query, itemId)

	return err
}

func (r *ItemsRepo) DeleteImages(ctx context.Context, itemId int) error {
	query := fmt.Sprintf("DELETE FROM %s WHERE item_id=$1;", itemsImagesTable)
	_, err := r.db.Exec(query, itemId)

	return err
}

func (r *ItemsRepo) DeleteColors(ctx context.Context, itemId int) error {
	query := fmt.Sprintf("DELETE FROM %s WHERE item_id=$1;", itemsColorsTable)
	_, err := r.db.Exec(query, itemId)

	return err
}

func (r *ItemsRepo) Exist(ctx context.Context, itemId int) (bool, error) {
	var exist bool
	queryMain := fmt.Sprintf("SELECT * FROM %s WHERE id=$1", itemsTable)
	query := fmt.Sprintf("SELECT exists (%s)", queryMain)
	if err := r.db.QueryRow(query, itemId).Scan(&exist); err != nil && err != sql.ErrNoRows {
		return false, err
	}
	return exist, nil
}
