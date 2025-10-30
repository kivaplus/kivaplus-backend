package repository

// PostgresRepository implements the usuario and pessoa repositories using PostgreSQL
// type PostgresRepository struct {
// 	db *database.Connection
// }

// // NewPostgresRepository creates a new PostgreSQL repository
// func NewPostgresRepository(db *database.Connection) *PostgresRepository {
// 	return &PostgresRepository{
// 		db: db,
// 	}
// }

// // Usuario Repository Methods

// // Create creates a new user in the database
// func (r *PostgresRepository) Create(ctx context.Context, usuario *domain.Usuario) error {
// 	query := `
// 		INSERT INTO usuario (email, senha_hash, ativo, created_at, updated_at)
// 		VALUES ($1, $2, $3, $4, $5)
// 		RETURNING id`

// 	err := r.db.DB.QueryRowContext(
// 		ctx,
// 		query,
// 		usuario.Email,
// 		usuario.SenhaHash,
// 		usuario.Ativo,
// 		usuario.CreatedAt,
// 		usuario.UpdatedAt,
// 	).Scan(&usuario.ID)

// 	if err != nil {
// 		return fmt.Errorf("failed to create user: %w", err)
// 	}

// 	return nil
// }

// // GetByEmail retrieves a user by email
// func (r *PostgresRepository) GetByEmail(ctx context.Context, email string) (*domain.Usuario, error) {
// 	query := `
// 		SELECT id, email, senha_hash, ativo, ultimo_acesso, pessoa_id, created_at, updated_at, deleted_at
// 		FROM usuario
// 		WHERE email = $1 AND deleted_at IS NULL`

// 	usuario := &domain.Usuario{}
// 	var ultimoAcesso, deletedAt sql.NullTime
// 	var pessoaID sql.NullInt64

// 	err := r.db.DB.QueryRowContext(ctx, query, email).Scan(
// 		&usuario.ID,
// 		&usuario.Email,
// 		&usuario.SenhaHash,
// 		&usuario.Ativo,
// 		&ultimoAcesso,
// 		&pessoaID,
// 		&usuario.CreatedAt,
// 		&usuario.UpdatedAt,
// 		&deletedAt,
// 	)

// 	if err != nil {
// 		if err == sql.ErrNoRows {
// 			return nil, fmt.Errorf("user not found")
// 		}
// 		return nil, fmt.Errorf("failed to get user by email: %w", err)
// 	}

// 	// Handle nullable fields
// 	if ultimoAcesso.Valid {
// 		usuario.UltimoAcesso = &ultimoAcesso.Time
// 	}
// 	if pessoaID.Valid {
// 		usuario.PessoaID = &pessoaID.Int64
// 	}
// 	if deletedAt.Valid {
// 		usuario.DeletedAt = &deletedAt.Time
// 	}

// 	return usuario, nil
// }

// // GetByID retrieves a user by ID
// func (r *PostgresRepository) GetByID(ctx context.Context, id int64) (*domain.Usuario, error) {
// 	query := `
// 		SELECT id, email, senha_hash, ativo, ultimo_acesso, pessoa_id, created_at, updated_at, deleted_at
// 		FROM usuario
// 		WHERE id = $1 AND deleted_at IS NULL`

// 	usuario := &domain.Usuario{}
// 	var ultimoAcesso, deletedAt sql.NullTime
// 	var pessoaID sql.NullInt64

// 	err := r.db.DB.QueryRowContext(ctx, query, id).Scan(
// 		&usuario.ID,
// 		&usuario.Email,
// 		&usuario.SenhaHash,
// 		&usuario.Ativo,
// 		&ultimoAcesso,
// 		&pessoaID,
// 		&usuario.CreatedAt,
// 		&usuario.UpdatedAt,
// 		&deletedAt,
// 	)

// 	if err != nil {
// 		if err == sql.ErrNoRows {
// 			return nil, fmt.Errorf("user not found")
// 		}
// 		return nil, fmt.Errorf("failed to get user by ID: %w", err)
// 	}

// 	// Handle nullable fields
// 	if ultimoAcesso.Valid {
// 		usuario.UltimoAcesso = &ultimoAcesso.Time
// 	}
// 	if pessoaID.Valid {
// 		usuario.PessoaID = &pessoaID.Int64
// 	}
// 	if deletedAt.Valid {
// 		usuario.DeletedAt = &deletedAt.Time
// 	}

// 	return usuario, nil
// }

// // Update updates an existing user
// func (r *PostgresRepository) Update(ctx context.Context, usuario *domain.Usuario) error {
// 	query := `
// 		UPDATE usuario
// 		SET email = $2, senha_hash = $3, ativo = $4, ultimo_acesso = $5, pessoa_id = $6, updated_at = $7
// 		WHERE id = $1 AND deleted_at IS NULL`

// 	_, err := r.db.DB.ExecContext(
// 		ctx,
// 		query,
// 		usuario.ID,
// 		usuario.Email,
// 		usuario.SenhaHash,
// 		usuario.Ativo,
// 		usuario.UltimoAcesso,
// 		usuario.PessoaID,
// 		time.Now(),
// 	)

// 	if err != nil {
// 		return fmt.Errorf("failed to update user: %w", err)
// 	}

// 	return nil
// }

// // Delete soft deletes a user
// func (r *PostgresRepository) Delete(ctx context.Context, id int64) error {
// 	query := `UPDATE usuario SET deleted_at = $2, updated_at = $2 WHERE id = $1`

// 	_, err := r.db.DB.ExecContext(ctx, query, id, time.Now())
// 	if err != nil {
// 		return fmt.Errorf("failed to delete user: %w", err)
// 	}

// 	return nil
// }

// // Exists checks if a user exists by email
// func (r *PostgresRepository) Exists(ctx context.Context, email string) (bool, error) {
// 	query := `SELECT EXISTS(SELECT 1 FROM usuario WHERE email = $1 AND deleted_at IS NULL)`

// 	var exists bool
// 	err := r.db.DB.QueryRowContext(ctx, query, email).Scan(&exists)
// 	if err != nil {
// 		return false, fmt.Errorf("failed to check if user exists: %w", err)
// 	}

// 	return exists, nil
// }

// // UpdateLastAccess updates the user's last access time
// func (r *PostgresRepository) UpdateLastAccess(ctx context.Context, id int64) error {
// 	query := `UPDATE usuario SET ultimo_acesso = $2, updated_at = $2 WHERE id = $1`

// 	_, err := r.db.DB.ExecContext(ctx, query, id, time.Now())
// 	if err != nil {
// 		return fmt.Errorf("failed to update last access: %w", err)
// 	}

// 	return nil
// }

// // Pessoa Repository Methods

// // Create creates a new person in the database
// func (r *PostgresRepository) Create(ctx context.Context, pessoa *domain.Pessoa) error {
// 	query := `
// 		INSERT INTO pessoa (nome, tipo_pessoa, telefone, email, endereco_id, profissao, estado_civil, data_nascimento, usuario_id, observacoes, created_at, updated_at)
// 		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12)
// 		RETURNING id`

// 	err := r.db.DB.QueryRowContext(
// 		ctx,
// 		query,
// 		pessoa.Nome,
// 		pessoa.TipoPessoa,
// 		pessoa.Telefone,
// 		pessoa.Email,
// 		pessoa.EnderecoID,
// 		pessoa.Profissao,
// 		pessoa.EstadoCivil,
// 		pessoa.DataNascimento,
// 		pessoa.UsuarioID,
// 		pessoa.Observacoes,
// 		pessoa.CreatedAt,
// 		pessoa.UpdatedAt,
// 	).Scan(&pessoa.ID)

// 	if err != nil {
// 		return fmt.Errorf("failed to create pessoa: %w", err)
// 	}

// 	return nil
// }

// // GetByID retrieves a person by ID
// func (r *PostgresRepository) GetByID(ctx context.Context, id int64) (*domain.Pessoa, error) {
// 	query := `
// 		SELECT id, nome, tipo_pessoa, telefone, email, endereco_id, profissao, estado_civil, data_nascimento, usuario_id, observacoes, created_at, updated_at, deleted_at
// 		FROM pessoa
// 		WHERE id = $1 AND deleted_at IS NULL`

// 	pessoa := &domain.Pessoa{}
// 	var telefone, email, profissao, observacoes sql.NullString
// 	var estadoCivil sql.NullString
// 	var dataNascimento, deletedAt sql.NullTime
// 	var usuarioID sql.NullInt64

// 	err := r.db.DB.QueryRowContext(ctx, query, id).Scan(
// 		&pessoa.ID,
// 		&pessoa.Nome,
// 		&pessoa.TipoPessoa,
// 		&telefone,
// 		&email,
// 		&pessoa.EnderecoID,
// 		&profissao,
// 		&estadoCivil,
// 		&dataNascimento,
// 		&usuarioID,
// 		&observacoes,
// 		&pessoa.CreatedAt,
// 		&pessoa.UpdatedAt,
// 		&deletedAt,
// 	)

// 	if err != nil {
// 		if err == sql.ErrNoRows {
// 			return nil, fmt.Errorf("pessoa not found")
// 		}
// 		return nil, fmt.Errorf("failed to get pessoa by ID: %w", err)
// 	}

// 	// Handle nullable fields
// 	if telefone.Valid {
// 		pessoa.Telefone = &telefone.String
// 	}
// 	if email.Valid {
// 		pessoa.Email = &email.String
// 	}
// 	if profissao.Valid {
// 		pessoa.Profissao = &profissao.String
// 	}
// 	if observacoes.Valid {
// 		pessoa.Observacoes = &observacoes.String
// 	}
// 	if estadoCivil.Valid {
// 		ec := domain.EstadoCivil(estadoCivil.String)
// 		pessoa.EstadoCivil = &ec
// 	}
// 	if dataNascimento.Valid {
// 		pessoa.DataNascimento = &dataNascimento.Time
// 	}
// 	if usuarioID.Valid {
// 		pessoa.UsuarioID = &usuarioID.Int64
// 	}
// 	if deletedAt.Valid {
// 		pessoa.DeletedAt = &deletedAt.Time
// 	}

// 	return pessoa, nil
// }

// // GetByUsuarioID retrieves a person by user ID
// func (r *PostgresRepository) GetByUsuarioID(ctx context.Context, usuarioID int64) (*domain.Pessoa, error) {
// 	query := `
// 		SELECT id, nome, tipo_pessoa, telefone, email, endereco_id, profissao, estado_civil, data_nascimento, usuario_id, observacoes, created_at, updated_at, deleted_at
// 		FROM pessoa
// 		WHERE usuario_id = $1 AND deleted_at IS NULL`

// 	pessoa := &domain.Pessoa{}
// 	var telefone, email, profissao, observacoes sql.NullString
// 	var estadoCivil sql.NullString
// 	var dataNascimento, deletedAt sql.NullTime
// 	var userID sql.NullInt64

// 	err := r.db.DB.QueryRowContext(ctx, query, usuarioID).Scan(
// 		&pessoa.ID,
// 		&pessoa.Nome,
// 		&pessoa.TipoPessoa,
// 		&telefone,
// 		&email,
// 		&pessoa.EnderecoID,
// 		&profissao,
// 		&estadoCivil,
// 		&dataNascimento,
// 		&userID,
// 		&observacoes,
// 		&pessoa.CreatedAt,
// 		&pessoa.UpdatedAt,
// 		&deletedAt,
// 	)

// 	if err != nil {
// 		if err == sql.ErrNoRows {
// 			return nil, fmt.Errorf("pessoa not found")
// 		}
// 		return nil, fmt.Errorf("failed to get pessoa by usuario ID: %w", err)
// 	}

// 	// Handle nullable fields (same as above)
// 	if telefone.Valid {
// 		pessoa.Telefone = &telefone.String
// 	}
// 	if email.Valid {
// 		pessoa.Email = &email.String
// 	}
// 	if profissao.Valid {
// 		pessoa.Profissao = &profissao.String
// 	}
// 	if observacoes.Valid {
// 		pessoa.Observacoes = &observacoes.String
// 	}
// 	if estadoCivil.Valid {
// 		ec := domain.EstadoCivil(estadoCivil.String)
// 		pessoa.EstadoCivil = &ec
// 	}
// 	if dataNascimento.Valid {
// 		pessoa.DataNascimento = &dataNascimento.Time
// 	}
// 	if userID.Valid {
// 		pessoa.UsuarioID = &userID.Int64
// 	}
// 	if deletedAt.Valid {
// 		pessoa.DeletedAt = &deletedAt.Time
// 	}

// 	return pessoa, nil
// }

// // Update updates an existing person
// func (r *PostgresRepository) Update(ctx context.Context, pessoa *domain.Pessoa) error {
// 	query := `
// 		UPDATE pessoa
// 		SET nome = $2, tipo_pessoa = $3, telefone = $4, email = $5, endereco_id = $6, profissao = $7, estado_civil = $8, data_nascimento = $9, usuario_id = $10, observacoes = $11, updated_at = $12
// 		WHERE id = $1 AND deleted_at IS NULL`

// 	_, err := r.db.DB.ExecContext(
// 		ctx,
// 		query,
// 		pessoa.ID,
// 		pessoa.Nome,
// 		pessoa.TipoPessoa,
// 		pessoa.Telefone,
// 		pessoa.Email,
// 		pessoa.EnderecoID,
// 		pessoa.Profissao,
// 		pessoa.EstadoCivil,
// 		pessoa.DataNascimento,
// 		pessoa.UsuarioID,
// 		pessoa.Observacoes,
// 		time.Now(),
// 	)

// 	if err != nil {
// 		return fmt.Errorf("failed to update pessoa: %w", err)
// 	}

// 	return nil
// }

// // Delete soft deletes a person
// func (r *PostgresRepository) Delete(ctx context.Context, id int64) error {
// 	query := `UPDATE pessoa SET deleted_at = $2, updated_at = $2 WHERE id = $1`

// 	_, err := r.db.DB.ExecContext(ctx, query, id, time.Now())
// 	if err != nil {
// 		return fmt.Errorf("failed to delete pessoa: %w", err)
// 	}

// 	return nil
// }
