package crud

// MongoRepository MongoDB仓储实现
// type MongoRepository[T ICrudEntity] struct {
// 	coll       *mongo.Collection
// 	entityType reflect.Type
// 	filters    []bson.M
// 	pipeline   []bson.M
// 	SoftDelete bool // 是否启用软删除
// }

// // NewMongoRepository 创建MongoDB仓储实例
// func NewMongoRepository[T ICrudEntity](db *mongo.Database, entity T) *MongoRepository[T] {
// 	entityType := reflect.TypeOf(entity)
// 	if entityType.Kind() == reflect.Ptr {
// 		entityType = entityType.Elem()
// 	}
// 	return &MongoRepository[T]{
// 		coll:       db.Collection(entity.TableName()),
// 		entityType: entityType,
// 		filters:    make([]bson.M, 0),
// 		pipeline:   make([]bson.M, 0),
// 	}
// }

// // Session 创建新会话
// func (r *MongoRepository[T]) Session() IRepository[T] {
// 	return &MongoRepository[T]{
// 		coll:       r.coll,
// 		entityType: r.entityType,
// 		filters:    make([]bson.M, 0),
// 		pipeline:   make([]bson.M, 0),
// 	}
// }

// // AddQueryHook 添加查询钩子
// func (r *MongoRepository[T]) AddQueryHook(hook QueryHook) IRepository[T] {
// 	// MongoDB暂不支持查询钩子
// 	return r
// }

// // Create 创建实体
// func (r *MongoRepository[T]) Create(ctx context.Context, entity T) error {
// 	_, err := r.coll.InsertOne(ctx, entity)
// 	return err
// }

// // Update 更新实体
// func (r *MongoRepository[T]) Update(ctx context.Context, entity T, updateFields map[string]interface{}) error {
// 	filter := bson.M{"_id": entity.GetID()}
// 	update := bson.M{"$set": updateFields}
// 	if updateFields == nil {
// 		update = bson.M{"$set": entity}
// 	}
// 	_, err := r.coll.UpdateOne(ctx, filter, update)
// 	return err
// }

// // Delete 删除实体
// func (r *MongoRepository[T]) Delete(ctx context.Context, entity T, opts ...*options.DeleteOptions) error {
// 	filter := bson.M{"_id": entity.GetID()}
// 	if r.SoftDelete == false {
// 		_, err := r.coll.DeleteOne(ctx, filter)
// 		return err
// 	}

// 	update := bson.M{"$set": bson.M{"deleted_at": time.Now()}}
// 	_, err := r.coll.UpdateOne(ctx, filter, update)
// 	return err
// }

// // DeleteById 根据ID删除
// func (r *MongoRepository[T]) DeleteById(ctx context.Context, id any, opts ...*options.DeleteOptions) error {
// 	entity := NewModel[T]()
// 	if err := entity.SetID(id); err != nil {
// 		return err
// 	}
// 	return r.Delete(ctx, entity, opts...)
// }

// // FindById 根据ID查询
// func (r *MongoRepository[T]) FindById(ctx context.Context, id any) (T, error) {
// 	entity := NewModel[T]()
// 	filter := bson.M{"_id": id, "deleted_at": nil}
// 	err := r.coll.FindOne(ctx, filter).Decode(&entity)
// 	return entity, err
// }

// // Find 查询实体列表
// func (r *MongoRepository[T]) Find(ctx context.Context, entity T, opts *options.FindOptions) ([]T, error) {
// 	var entities []T
// 	filter := bson.M{"deleted_at": nil}

// 	findOptions := options.Find()
// 	// if opts != nil {
// 	// 	if opts.Page > 0 && opts.PageSize > 0 {
// 	// 		findOptions.SetSkip(int64((opts.Page - 1) * opts.PageSize))
// 	// 		findOptions.SetLimit(int64(opts.PageSize))
// 	// 	}
// 	// 	if opts.OrderBy != "" {
// 	// 		findOptions.SetSort(bson.D{{opts.OrderBy, opts.Order}})
// 	// 	}
// 	// }

// 	cursor, err := r.coll.Find(ctx, filter, findOptions)
// 	if err != nil {
// 		return nil, err
// 	}
// 	defer cursor.Close(ctx)

// 	if err = cursor.All(ctx, &entities); err != nil {
// 		return nil, err
// 	}
// 	return entities, nil
// }

// // Count 统计记录数
// func (r *MongoRepository[T]) Count(ctx context.Context, entity T) (int64, error) {
// 	filter := bson.M{"deleted_at": nil}
// 	return r.coll.CountDocuments(ctx, filter)
// }

// // BatchCreate 批量创建
// func (r *MongoRepository[T]) BatchCreate(ctx context.Context, entities []T, opts ...*options.BatchOptions) error {
// 	docs := make([]interface{}, len(entities))
// 	for i, entity := range entities {
// 		docs[i] = entity
// 	}
// 	_, err := r.coll.InsertMany(ctx, docs)
// 	return err
// }

// // BatchUpdate 批量更新
// func (r *MongoRepository[T]) BatchUpdate(ctx context.Context, entities []T) error {
// 	for _, entity := range entities {
// 		if err := r.Update(ctx, entity, nil); err != nil {
// 			return err
// 		}
// 	}
// 	return nil
// }

// // BatchDelete 批量删除
// func (r *MongoRepository[T]) BatchDelete(ctx context.Context, ids []any, opts ...*options.DeleteOptions) error {
// 	filter := bson.M{"_id": bson.M{"$in": ids}}
// 	if r.SoftDelete == false {
// 		_, err := r.coll.DeleteMany(ctx, filter)
// 		return err
// 	}

// 	update := bson.M{"$set": bson.M{"deleted_at": time.Now()}}
// 	_, err := r.coll.UpdateMany(ctx, filter, update)
// 	return err
// }

// // FindOne 查询单个实体
// func (r *MongoRepository[T]) FindOne(ctx context.Context, query interface{}, args ...interface{}) (T, error) {
// 	entity := NewModel[T]()
// 	filter := r.buildFilter(query, args...)
// 	filter["deleted_at"] = nil

// 	err := r.coll.FindOne(ctx, filter).Decode(&entity)
// 	return entity, err
// }

// // FindAll 查询所有符合条件的记录
// func (r *MongoRepository[T]) FindAll(ctx context.Context, query interface{}, args ...interface{}) ([]T, error) {
// 	var entities []T
// 	filter := r.buildFilter(query, args...)
// 	filter["deleted_at"] = nil

// 	cursor, err := r.coll.Find(ctx, filter)
// 	if err != nil {
// 		return nil, err
// 	}
// 	defer cursor.Close(ctx)

// 	if err = cursor.All(ctx, &entities); err != nil {
// 		return nil, err
// 	}
// 	return entities, nil
// }

// // Exists 检查记录是否存在
// func (r *MongoRepository[T]) Exists(ctx context.Context, query interface{}, args ...interface{}) (bool, error) {
// 	filter := r.buildFilter(query, args...)
// 	filter["deleted_at"] = nil

// 	count, err := r.coll.CountDocuments(ctx, filter)
// 	return count > 0, err
// }

// // Page 分页查询
// func (r *MongoRepository[T]) Page(ctx context.Context, page int, pageSize int) ([]T, int64, error) {
// 	var entities []T
// 	filter := bson.M{"deleted_at": nil}

// 	// 获取总数
// 	total, err := r.coll.CountDocuments(ctx, filter)
// 	if err != nil {
// 		return nil, 0, err
// 	}

// 	// 分页查询
// 	opts := options.Find().SetSkip(int64((page - 1) * pageSize)).SetLimit(int64(pageSize))
// 	cursor, err := r.coll.Find(ctx, filter, opts)
// 	if err != nil {
// 		return nil, 0, err
// 	}
// 	defer cursor.Close(ctx)

// 	if err = cursor.All(ctx, &entities); err != nil {
// 		return nil, 0, err
// 	}

// 	return entities, total, nil
// }

// // Where 条件查询
// func (r *MongoRepository[T]) Where(query interface{}, args ...interface{}) IRepository[T] {
// 	filter := r.buildFilter(query, args...)
// 	r.filters = append(r.filters, filter)
// 	return r
// }

// // Order 排序
// func (r *MongoRepository[T]) Order(value interface{}) IRepository[T] {
// 	r.pipeline = append(r.pipeline, bson.M{"$sort": value})
// 	return r
// }

// // Select 选择字段
// func (r *MongoRepository[T]) Select(query interface{}, args ...interface{}) IRepository[T] {
// 	fields := make(bson.M)
// 	switch v := query.(type) {
// 	case string:
// 		fields[v] = 1
// 	case []string:
// 		for _, field := range v {
// 			fields[field] = 1
// 		}
// 	}
// 	r.pipeline = append(r.pipeline, bson.M{"$project": fields})
// 	return r
// }

// // Preload 预加载关联
// func (r *MongoRepository[T]) Preload(query ...string) IRepository[T] {
// 	// MongoDB使用聚合管道实现关联查询
// 	for _, field := range query {
// 		r.pipeline = append(r.pipeline, bson.M{"$lookup": bson.M{
// 			"from":         field,
// 			"localField":   field + "_id",
// 			"foreignField": "_id",
// 			"as":           field,
// 		}})
// 	}
// 	return r
// }

// // Joins 连接查询
// func (r *MongoRepository[T]) Joins(query string, args ...interface{}) IRepository[T] {
// 	// MongoDB使用$lookup实现连接查询
// 	return r
// }

// // Group 分组
// func (r *MongoRepository[T]) Group(query string) IRepository[T] {
// 	r.pipeline = append(r.pipeline, bson.M{"$group": bson.M{"_id": "$" + query}})
// 	return r
// }

// // Having 过滤条件
// func (r *MongoRepository[T]) Having(query interface{}, args ...interface{}) IRepository[T] {
// 	filter := r.buildFilter(query, args...)
// 	r.pipeline = append(r.pipeline, bson.M{"$match": filter})
// 	return r
// }

// // Transaction 事务操作
// func (r *MongoRepository[T]) Transaction(ctx context.Context, fc func(tx IRepository[T]) error) error {
// 	session, err := r.coll.Database().Client().StartSession()
// 	if err != nil {
// 		return err
// 	}
// 	defer session.EndSession(ctx)

// 	_, err = session.WithTransaction(ctx, func(sessCtx mongo.SessionContext) (interface{}, error) {
// 		return nil, fc(r)
// 	})
// 	return err
// }

// // WithTx 使用事务
// func (r *MongoRepository[T]) WithTx(tx *gorm.DB) IRepository[T] {
// 	// MongoDB事务需要使用session，这里暂不支持
// 	return r
// }

// // LockForUpdate 行锁
// func (r *MongoRepository[T]) LockForUpdate() IRepository[T] {
// 	// MongoDB不支持行锁，可以使用事务实现
// 	return r
// }

// // SharedLock 共享锁
// func (r *MongoRepository[T]) SharedLock() IRepository[T] {
// 	// MongoDB不支持共享锁
// 	return r
// }

// // 聚合操作
// func (r *MongoRepository[T]) Sum(ctx context.Context, field string) (float64, error) {
// 	pipeline := []bson.M{
// 		{"$match": bson.M{"deleted_at": nil}},
// 		{"$group": bson.M{
// 			"_id":   nil,
// 			"total": bson.M{"$sum": "$" + field},
// 		}},
// 	}

// 	cursor, err := r.coll.Aggregate(ctx, pipeline)
// 	if err != nil {
// 		return 0, err
// 	}
// 	defer cursor.Close(ctx)

// 	var result []bson.M
// 	if err = cursor.All(ctx, &result); err != nil {
// 		return 0, err
// 	}

// 	if len(result) == 0 {
// 		return 0, nil
// 	}
// 	return result[0]["total"].(float64), nil
// }

// func (r *MongoRepository[T]) CountField(ctx context.Context, field string) (int64, error) {
// 	pipeline := []bson.M{
// 		{"$match": bson.M{"deleted_at": nil}},
// 		{"$group": bson.M{
// 			"_id":   nil,
// 			"count": bson.M{"$sum": 1},
// 		}},
// 	}

// 	cursor, err := r.coll.Aggregate(ctx, pipeline)
// 	if err != nil {
// 		return 0, err
// 	}
// 	defer cursor.Close(ctx)

// 	var result []bson.M
// 	if err = cursor.All(ctx, &result); err != nil {
// 		return 0, err
// 	}

// 	if len(result) == 0 {
// 		return 0, nil
// 	}
// 	return result[0]["count"].(int64), nil
// }

// func (r *MongoRepository[T]) Max(ctx context.Context, field string) (float64, error) {
// 	pipeline := []bson.M{
// 		{"$match": bson.M{"deleted_at": nil}},
// 		{"$group": bson.M{
// 			"_id": nil,
// 			"max": bson.M{"$max": "$" + field},
// 		}},
// 	}

// 	cursor, err := r.coll.Aggregate(ctx, pipeline)
// 	if err != nil {
// 		return 0, err
// 	}
// 	defer cursor.Close(ctx)

// 	var result []bson.M
// 	if err = cursor.All(ctx, &result); err != nil {
// 		return 0, err
// 	}

// 	if len(result) == 0 {
// 		return 0, nil
// 	}
// 	return result[0]["max"].(float64), nil
// }

// func (r *MongoRepository[T]) Min(ctx context.Context, field string) (float64, error) {
// 	pipeline := []bson.M{
// 		{"$match": bson.M{"deleted_at": nil}},
// 		{"$group": bson.M{
// 			"_id": nil,
// 			"min": bson.M{"$min": "$" + field},
// 		}},
// 	}

// 	cursor, err := r.coll.Aggregate(ctx, pipeline)
// 	if err != nil {
// 		return 0, err
// 	}
// 	defer cursor.Close(ctx)

// 	var result []bson.M
// 	if err = cursor.All(ctx, &result); err != nil {
// 		return 0, err
// 	}

// 	if len(result) == 0 {
// 		return 0, nil
// 	}
// 	return result[0]["min"].(float64), nil
// }

// func (r *MongoRepository[T]) Avg(ctx context.Context, field string) (float64, error) {
// 	pipeline := []bson.M{
// 		{"$match": bson.M{"deleted_at": nil}},
// 		{"$group": bson.M{
// 			"_id": nil,
// 			"avg": bson.M{"$avg": "$" + field},
// 		}},
// 	}

// 	cursor, err := r.coll.Aggregate(ctx, pipeline)
// 	if err != nil {
// 		return 0, err
// 	}
// 	defer cursor.Close(ctx)

// 	var result []bson.M
// 	if err = cursor.All(ctx, &result); err != nil {
// 		return 0, err
// 	}

// 	if len(result) == 0 {
// 		return 0, nil
// 	}
// 	return result[0]["avg"].(float64), nil
// }

// // buildFilter 构建查询条件
// func (r *MongoRepository[T]) buildFilter(query interface{}, args ...interface{}) bson.M {
// 	filter := bson.M{}

// 	switch v := query.(type) {
// 	case string:
// 		// 简单字符串查询
// 		if len(args) > 0 {
// 			filter[v] = args[0]
// 		}
// 	case map[string]interface{}:
// 		// 直接使用map作为查询条件
// 		for k, val := range v {
// 			filter[k] = val
// 		}
// 	case bson.M:
// 		// 直接使用bson.M作为查询条件
// 		filter = v
// 	}

// 	return filter
// }
