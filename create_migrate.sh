#!/bin/bash
# Script para criar a estrutura de migrations

# Criar diretório de migrations
mkdir -p internal/database/migrations

# ============================================================
# 000001_create_function.up.sql
# ============================================================
cat > internal/database/migrations/000001_create_function.up.sql << 'EOF'
CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ language 'plpgsql';
EOF

cat > internal/database/migrations/000001_create_function.down.sql << 'EOF'
DROP FUNCTION IF EXISTS update_updated_at_column();
EOF

# ============================================================
# 000002_create_endereco.up.sql
# ============================================================
cat > internal/database/migrations/000002_create_endereco.up.sql << 'EOF'
CREATE TABLE endereco (
    id BIGSERIAL PRIMARY KEY,
    cep VARCHAR(10) NOT NULL,
    logradouro VARCHAR(255) NOT NULL,
    numero VARCHAR(20) NOT NULL,
    complemento VARCHAR(255),
    bairro VARCHAR(100) NOT NULL,
    cidade VARCHAR(100) NOT NULL,
    estado VARCHAR(2) NOT NULL,
    pais VARCHAR(2) NOT NULL DEFAULT 'BR',
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_endereco_cep ON endereco(cep);
CREATE INDEX idx_endereco_cidade_estado ON endereco(cidade, estado);

CREATE TRIGGER update_endereco_updated_at
    BEFORE UPDATE ON endereco
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();
EOF

cat > internal/database/migrations/000002_create_endereco.down.sql << 'EOF'
DROP TRIGGER IF EXISTS update_endereco_updated_at ON endereco;
DROP TABLE IF EXISTS endereco CASCADE;
EOF

# ============================================================
# 000003_create_pessoa.up.sql
# ============================================================
cat > internal/database/migrations/000003_create_pessoa.up.sql << 'EOF'
CREATE TABLE pessoa (
    id BIGSERIAL PRIMARY KEY,
    nome VARCHAR(255) NOT NULL,
    tipo_pessoa VARCHAR(2) NOT NULL CHECK (tipo_pessoa IN ('PF', 'PJ')),
    telefone VARCHAR(20),
    email VARCHAR(255),
    endereco_id BIGINT NOT NULL REFERENCES endereco(id) ON DELETE RESTRICT,
    profissao VARCHAR(100),
    estado_civil VARCHAR(20) CHECK (estado_civil IN ('solteiro', 'casado', 'divorciado', 'viuvo', 'uniao_estavel', 'outro')),
    data_nascimento DATE,
    usuario_id BIGINT,
    observacoes TEXT,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMP
);

CREATE INDEX idx_pessoa_nome ON pessoa(nome);
CREATE INDEX idx_pessoa_email ON pessoa(email);
CREATE INDEX idx_pessoa_tipo ON pessoa(tipo_pessoa);
CREATE INDEX idx_pessoa_deleted_at ON pessoa(deleted_at);

CREATE TRIGGER update_pessoa_updated_at
    BEFORE UPDATE ON pessoa
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();
EOF

cat > internal/database/migrations/000003_create_pessoa.down.sql << 'EOF'
DROP TRIGGER IF EXISTS update_pessoa_updated_at ON pessoa;
DROP TABLE IF EXISTS pessoa CASCADE;
EOF

# ============================================================
# 000004_create_documento.up.sql
# ============================================================
cat > internal/database/migrations/000004_create_documento.up.sql << 'EOF'
CREATE TABLE documento (
    id BIGSERIAL PRIMARY KEY,
    pessoa_id BIGINT NOT NULL REFERENCES pessoa(id) ON DELETE CASCADE,
    tipo VARCHAR(10) NOT NULL CHECK (tipo IN ('CPF', 'RG', 'CNH', 'CNPJ')),
    numero VARCHAR(30) NOT NULL,
    orgao_emissor VARCHAR(20),
    data_emissao DATE,
    data_validade DATE,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW(),
    CONSTRAINT uk_pessoa_tipo_documento UNIQUE(pessoa_id, tipo)
);

CREATE INDEX idx_documento_pessoa_id ON documento(pessoa_id);
CREATE INDEX idx_documento_numero ON documento(numero);

CREATE TRIGGER update_documento_updated_at
    BEFORE UPDATE ON documento
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();
EOF

cat > internal/database/migrations/000004_create_documento.down.sql << 'EOF'
DROP TRIGGER IF EXISTS update_documento_updated_at ON documento;
DROP TABLE IF EXISTS documento CASCADE;
EOF

# ============================================================
# 000005_create_usuario.up.sql
# ============================================================
cat > internal/database/migrations/000005_create_usuario.up.sql << 'EOF'
CREATE TABLE usuario (
    id BIGSERIAL PRIMARY KEY,
    email VARCHAR(255) NOT NULL UNIQUE,
    senha_hash VARCHAR(255) NOT NULL,
    ativo BOOLEAN NOT NULL DEFAULT TRUE,
    ultimo_acesso TIMESTAMP,
    pessoa_id BIGINT REFERENCES pessoa(id) ON DELETE SET NULL,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMP
);

CREATE INDEX idx_usuario_email ON usuario(email);
CREATE INDEX idx_usuario_ativo ON usuario(ativo);
CREATE INDEX idx_usuario_pessoa_id ON usuario(pessoa_id);

CREATE TRIGGER update_usuario_updated_at
    BEFORE UPDATE ON usuario
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();

ALTER TABLE pessoa
    ADD CONSTRAINT fk_pessoa_usuario
    FOREIGN KEY (usuario_id) REFERENCES usuario(id) ON DELETE SET NULL;

CREATE INDEX idx_pessoa_usuario_id ON pessoa(usuario_id);
EOF

cat > internal/database/migrations/000005_create_usuario.down.sql << 'EOF'
ALTER TABLE pessoa DROP CONSTRAINT IF EXISTS fk_pessoa_usuario;
DROP TRIGGER IF EXISTS update_usuario_updated_at ON usuario;
DROP TABLE IF EXISTS usuario CASCADE;
EOF

# ============================================================
# 000006_create_papel.up.sql
# ============================================================
cat > internal/database/migrations/000006_create_papel.up.sql << 'EOF'
CREATE TABLE papel (
    id SERIAL PRIMARY KEY,
    nome VARCHAR(50) NOT NULL UNIQUE,
    descricao TEXT,
    permissoes JSONB NOT NULL DEFAULT '{}',
    created_at TIMESTAMP NOT NULL DEFAULT NOW()
);

INSERT INTO papel (nome, descricao, permissoes) VALUES
('super_admin', 'Administrador do sistema - acesso total', '{"all": true}'),
('sindico', 'Síndico do condomínio - gestão completa do condomínio', '{
    "condominio": {"read": true, "update": true},
    "unidade": {"create": true, "read": true, "update": true, "delete": true},
    "morador": {"create": true, "read": true, "update": true, "delete": true},
    "funcionario": {"create": true, "read": true, "update": true, "delete": true},
    "veiculo": {"create": true, "read": true, "update": true, "delete": true},
    "contrato": {"create": true, "read": true, "update": true, "delete": true}
}'),
('morador', 'Morador - acesso restrito aos próprios dados', '{
    "unidade": {"read": true},
    "morador": {"read": true},
    "veiculo": {"create": true, "read": true, "update": true}
}'),
('funcionario', 'Funcionário do condomínio', '{
    "morador": {"read": true},
    "veiculo": {"read": true}
}');
EOF

cat > internal/database/migrations/000006_create_papel.down.sql << 'EOF'
DROP TABLE IF EXISTS papel CASCADE;
EOF

# Continue criando os outros arquivos...
# (000007 até 000017 seguindo o mesmo padrão)

# ============================================================
# 000007_create_condominio.up.sql
# ============================================================
cat > internal/database/migrations/000007_create_condominio.up.sql << 'EOF'
CREATE TABLE condominio (
    id BIGSERIAL PRIMARY KEY,
    nome VARCHAR(255) NOT NULL,
    cnpj VARCHAR(18) NOT NULL UNIQUE,
    telefone VARCHAR(20),
    email VARCHAR(255),
    endereco_id BIGINT NOT NULL REFERENCES endereco(id) ON DELETE RESTRICT,
    sindico_id BIGINT NOT NULL REFERENCES pessoa(id) ON DELETE RESTRICT,
    administrador_id BIGINT REFERENCES pessoa(id) ON DELETE SET NULL,
    observacoes TEXT,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMP
);

CREATE INDEX idx_condominio_cnpj ON condominio(cnpj);
CREATE INDEX idx_condominio_nome ON condominio(nome);
CREATE INDEX idx_condominio_sindico_id ON condominio(sindico_id);
CREATE INDEX idx_condominio_deleted_at ON condominio(deleted_at);

CREATE TRIGGER update_condominio_updated_at
    BEFORE UPDATE ON condominio
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();
EOF

cat > internal/database/migrations/000007_create_condominio.down.sql << 'EOF'
DROP TRIGGER IF EXISTS update_condominio_updated_at ON condominio;
DROP TABLE IF EXISTS condominio CASCADE;
EOF

# ============================================================
# 000008_create_usuario_papel.up.sql
# ============================================================
cat > internal/database/migrations/000008_create_usuario_papel.up.sql << 'EOF'
CREATE TABLE usuario_papel (
    id BIGSERIAL PRIMARY KEY,
    usuario_id BIGINT NOT NULL REFERENCES usuario(id) ON DELETE CASCADE,
    papel_id INTEGER NOT NULL REFERENCES papel(id) ON DELETE CASCADE,
    condominio_id BIGINT REFERENCES condominio(id) ON DELETE CASCADE,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    CONSTRAINT uk_usuario_papel_condominio UNIQUE(usuario_id, papel_id, condominio_id)
);

CREATE INDEX idx_usuario_papel_usuario ON usuario_papel(usuario_id);
CREATE INDEX idx_usuario_papel_papel ON usuario_papel(papel_id);
CREATE INDEX idx_usuario_papel_condominio ON usuario_papel(condominio_id);
EOF

cat > internal/database/migrations/000008_create_usuario_papel.down.sql << 'EOF'
DROP TABLE IF EXISTS usuario_papel CASCADE;
EOF

# ============================================================
# 000009_create_unidade.up.sql
# ============================================================
cat > internal/database/migrations/000009_create_unidade.up.sql << 'EOF'
CREATE TABLE unidade (
    id BIGSERIAL PRIMARY KEY,
    condominio_id BIGINT NOT NULL REFERENCES condominio(id) ON DELETE CASCADE,
    numero VARCHAR(20) NOT NULL,
    bloco VARCHAR(10),
    andar INTEGER,
    area_m2 NUMERIC(10, 2),
    proprietario_id BIGINT NOT NULL REFERENCES pessoa(id) ON DELETE RESTRICT,
    tem_garagem BOOLEAN NOT NULL DEFAULT FALSE,
    observacoes TEXT,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMP,
    CONSTRAINT uk_condominio_bloco_numero UNIQUE(condominio_id, bloco, numero)
);

CREATE INDEX idx_unidade_condominio_id ON unidade(condominio_id);
CREATE INDEX idx_unidade_proprietario_id ON unidade(proprietario_id);
CREATE INDEX idx_unidade_numero ON unidade(numero);
CREATE INDEX idx_unidade_deleted_at ON unidade(deleted_at);

CREATE TRIGGER update_unidade_updated_at
    BEFORE UPDATE ON unidade
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();
EOF

cat > internal/database/migrations/000009_create_unidade.down.sql << 'EOF'
DROP TRIGGER IF EXISTS update_unidade_updated_at ON unidade;
DROP TABLE IF EXISTS unidade CASCADE;
EOF

# ============================================================
# 000010_create_morador.up.sql
# ============================================================
cat > internal/database/migrations/000010_create_morador.up.sql << 'EOF'
CREATE TABLE morador (
    id BIGSERIAL PRIMARY KEY,
    pessoa_id BIGINT NOT NULL UNIQUE REFERENCES pessoa(id) ON DELETE CASCADE,
    unidade_id BIGINT NOT NULL REFERENCES unidade(id) ON DELETE CASCADE,
    tipo VARCHAR(20) NOT NULL CHECK (tipo IN ('proprietario', 'inquilino', 'dependente')),
    status VARCHAR(20) NOT NULL DEFAULT 'ativo' CHECK (status IN ('ativo', 'inativo')),
    data_entrada DATE NOT NULL,
    data_saida DATE,
    responsavel_financeiro BOOLEAN NOT NULL DEFAULT FALSE,
    observacoes TEXT,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMP,
    CONSTRAINT chk_data_saida CHECK (data_saida IS NULL OR data_saida >= data_entrada)
);

CREATE INDEX idx_morador_pessoa_id ON morador(pessoa_id);
CREATE INDEX idx_morador_unidade_id ON morador(unidade_id);
CREATE INDEX idx_morador_status ON morador(status);
CREATE INDEX idx_morador_deleted_at ON morador(deleted_at);

CREATE TRIGGER update_morador_updated_at
    BEFORE UPDATE ON morador
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();
EOF

cat > internal/database/migrations/000010_create_morador.down.sql << 'EOF'
DROP TRIGGER IF EXISTS update_morador_updated_at ON morador;
DROP TABLE IF EXISTS morador CASCADE;
EOF

# ============================================================
# 000011_create_funcionario.up.sql
# ============================================================
cat > internal/database/migrations/000011_create_funcionario.up.sql << 'EOF'
CREATE TABLE funcionario (
    id BIGSERIAL PRIMARY KEY,
    pessoa_id BIGINT NOT NULL REFERENCES pessoa(id) ON DELETE CASCADE,
    condominio_id BIGINT NOT NULL REFERENCES condominio(id) ON DELETE CASCADE,
    cargo VARCHAR(100) NOT NULL,
    salario NUMERIC(12, 2),
    data_admissao DATE NOT NULL,
    data_demissao DATE,
    status VARCHAR(20) NOT NULL DEFAULT 'ativo' CHECK (status IN ('ativo', 'inativo', 'afastado', 'ferias')),
    observacoes TEXT,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMP,
    CONSTRAINT uk_pessoa_condominio_funcionario UNIQUE(pessoa_id, condominio_id),
    CONSTRAINT chk_data_demissao CHECK (data_demissao IS NULL OR data_demissao >= data_admissao)
);

CREATE INDEX idx_funcionario_pessoa_id ON funcionario(pessoa_id);
CREATE INDEX idx_funcionario_condominio_id ON funcionario(condominio_id);
CREATE INDEX idx_funcionario_status ON funcionario(status);
CREATE INDEX idx_funcionario_deleted_at ON funcionario(deleted_at);

CREATE TRIGGER update_funcionario_updated_at
    BEFORE UPDATE ON funcionario
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();
EOF

cat > internal/database/migrations/000011_create_funcionario.down.sql << 'EOF'
DROP TRIGGER IF EXISTS update_funcionario_updated_at ON funcionario;
DROP TABLE IF EXISTS funcionario CASCADE;
EOF

# ============================================================
# 000012_create_veiculo.up.sql
# ============================================================
cat > internal/database/migrations/000012_create_veiculo.up.sql << 'EOF'
CREATE TABLE veiculo (
    id BIGSERIAL PRIMARY KEY,
    placa VARCHAR(10) NOT NULL UNIQUE,
    modelo VARCHAR(100) NOT NULL,
    marca VARCHAR(100) NOT NULL,
    cor VARCHAR(50),
    ano INTEGER,
    tipo VARCHAR(20) NOT NULL DEFAULT 'carro' CHECK (tipo IN ('carro', 'moto', 'caminhonete', 'outros')),
    observacoes TEXT,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMP
);

CREATE INDEX idx_veiculo_placa ON veiculo(placa);
CREATE INDEX idx_veiculo_tipo ON veiculo(tipo);
CREATE INDEX idx_veiculo_deleted_at ON veiculo(deleted_at);

CREATE TRIGGER update_veiculo_updated_at
    BEFORE UPDATE ON veiculo
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();
EOF

cat > internal/database/migrations/000012_create_veiculo.down.sql << 'EOF'
DROP TRIGGER IF EXISTS update_veiculo_updated_at ON veiculo;
DROP TABLE IF EXISTS veiculo CASCADE;
EOF

# ============================================================
# 000013_create_morador_veiculo.up.sql
# ============================================================
cat > internal/database/migrations/000013_create_morador_veiculo.up.sql << 'EOF'
CREATE TABLE morador_veiculo (
    id BIGSERIAL PRIMARY KEY,
    morador_id BIGINT NOT NULL REFERENCES morador(id) ON DELETE CASCADE,
    veiculo_id BIGINT NOT NULL REFERENCES veiculo(id) ON DELETE CASCADE,
    principal BOOLEAN NOT NULL DEFAULT FALSE,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    CONSTRAINT uk_morador_veiculo UNIQUE(morador_id, veiculo_id)
);

CREATE INDEX idx_morador_veiculo_morador ON morador_veiculo(morador_id);
CREATE INDEX idx_morador_veiculo_veiculo ON morador_veiculo(veiculo_id);
EOF

cat > internal/database/migrations/000013_create_morador_veiculo.down.sql << 'EOF'
DROP TABLE IF EXISTS morador_veiculo CASCADE;
EOF

# ============================================================
# 000014_create_vaga_garagem.up.sql
# ============================================================
cat > internal/database/migrations/000014_create_vaga_garagem.up.sql << 'EOF'
CREATE TABLE vaga_garagem (
    id BIGSERIAL PRIMARY KEY,
    unidade_id BIGINT NOT NULL REFERENCES unidade(id) ON DELETE CASCADE,
    numero VARCHAR(20) NOT NULL,
    tipo VARCHAR(20) NOT NULL CHECK (tipo IN ('coberta', 'descoberta')),
    disponivel BOOLEAN NOT NULL DEFAULT TRUE,
    veiculo_id BIGINT REFERENCES veiculo(id) ON DELETE SET NULL,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW(),
    CONSTRAINT uk_unidade_numero_vaga UNIQUE(unidade_id, numero)
);

CREATE INDEX idx_vaga_garagem_unidade_id ON vaga_garagem(unidade_id);
CREATE INDEX idx_vaga_garagem_veiculo_id ON vaga_garagem(veiculo_id);
CREATE INDEX idx_vaga_garagem_disponivel ON vaga_garagem(disponivel);

CREATE TRIGGER update_vaga_garagem_updated_at
    BEFORE UPDATE ON vaga_garagem
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();
EOF

cat > internal/database/migrations/000014_create_vaga_garagem.down.sql << 'EOF'
DROP TRIGGER IF EXISTS update_vaga_garagem_updated_at ON vaga_garagem;
DROP TABLE IF EXISTS vaga_garagem CASCADE;
EOF

# ============================================================
# 000015_create_empresa_terceirizada.up.sql
# ============================================================
cat > internal/database/migrations/000015_create_empresa_terceirizada.up.sql << 'EOF'
CREATE TABLE empresa_terceirizada (
    id BIGSERIAL PRIMARY KEY,
    nome VARCHAR(255) NOT NULL,
    cnpj VARCHAR(18) NOT NULL UNIQUE,
    razao_social VARCHAR(255) NOT NULL,
    telefone VARCHAR(20),
    email VARCHAR(255),
    endereco_id BIGINT NOT NULL REFERENCES endereco(id) ON DELETE RESTRICT,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMP
);

CREATE INDEX idx_empresa_terceirizada_cnpj ON empresa_terceirizada(cnpj);
CREATE INDEX idx_empresa_terceirizada_nome ON empresa_terceirizada(nome);
CREATE INDEX idx_empresa_terceirizada_deleted_at ON empresa_terceirizada(deleted_at);

CREATE TRIGGER update_empresa_terceirizada_updated_at
    BEFORE UPDATE ON empresa_terceirizada
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();
EOF

cat > internal/database/migrations/000015_create_empresa_terceirizada.down.sql << 'EOF'
DROP TRIGGER IF EXISTS update_empresa_terceirizada_updated_at ON empresa_terceirizada;
DROP TABLE IF EXISTS empresa_terceirizada CASCADE;
EOF

# ============================================================
# 000016_create_contrato_terceirizado.up.sql
# ============================================================
cat > internal/database/migrations/000016_create_contrato_terceirizado.up.sql << 'EOF'
CREATE TABLE contrato_terceirizado (
    id BIGSERIAL PRIMARY KEY,
    condominio_id BIGINT NOT NULL REFERENCES condominio(id) ON DELETE CASCADE,
    empresa_id BIGINT NOT NULL REFERENCES empresa_terceirizada(id) ON DELETE RESTRICT,
    numero_contrato VARCHAR(100) NOT NULL UNIQUE,
    valor_mensal NUMERIC(12, 2) NOT NULL,
    dia_vencimento INTEGER NOT NULL CHECK (dia_vencimento >= 1 AND dia_vencimento <= 31),
    data_inicio DATE NOT NULL,
    data_fim DATE,
    status VARCHAR(20) NOT NULL DEFAULT 'ativo' CHECK (status IN ('ativo', 'suspenso', 'cancelado', 'finalizado')),
    observacoes TEXT,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMP,
    CONSTRAINT chk_data_fim CHECK (data_fim IS NULL OR data_fim >= data_inicio)
);

CREATE INDEX idx_contrato_terceirizado_condominio_id ON contrato_terceirizado(condominio_id);
CREATE INDEX idx_contrato_terceirizado_empresa_id ON contrato_terceirizado(empresa_id);
CREATE INDEX idx_contrato_terceirizado_status ON contrato_terceirizado(status);
CREATE INDEX idx_contrato_terceirizado_numero ON contrato_terceirizado(numero_contrato);
CREATE INDEX idx_contrato_terceirizado_deleted_at ON contrato_terceirizado(deleted_at);

CREATE TRIGGER update_contrato_terceirizado_updated_at
    BEFORE UPDATE ON contrato_terceirizado
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();
EOF

cat > internal/database/migrations/000016_create_contrato_terceirizado.down.sql << 'EOF'
DROP TRIGGER IF EXISTS update_contrato_terceirizado_updated_at ON contrato_terceirizado;
DROP TABLE IF EXISTS contrato_terceirizado CASCADE;
EOF

# ============================================================
# 000017_create_servico.up.sql
# ============================================================
cat > internal/database/migrations/000017_create_servico.up.sql << 'EOF'
CREATE TABLE servico (
    id BIGSERIAL PRIMARY KEY,
    contrato_id BIGINT NOT NULL REFERENCES contrato_terceirizado(id) ON DELETE CASCADE,
    nome VARCHAR(255) NOT NULL,
    descricao TEXT,
    valor NUMERIC(12, 2) NOT NULL,
    frequencia VARCHAR(20) NOT NULL CHECK (frequencia IN ('unico', 'diario', 'semanal', 'mensal', 'anual')),
    observacoes TEXT,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_servico_contrato_id ON servico(contrato_id);
CREATE INDEX idx_servico_nome ON servico(nome);

CREATE TRIGGER update_servico_updated_at
    BEFORE UPDATE ON servico
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();
EOF

cat > internal/database/migrations/000017_create_servico.down.sql << 'EOF'
DROP TRIGGER IF EXISTS update_servico_updated_at ON servico;
DROP TABLE IF EXISTS servico CASCADE;
EOF

echo "✅ Todos os arquivos de migration foram criados!"
echo "📁 Estrutura em: internal/database/migrations/"
