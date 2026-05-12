package database

import "gorm.io/gorm"

func Seed(db *gorm.DB) error {
	if err := seedProducts(db); err != nil {
		return err
	}
	return seedMaintenances(db)
}

func seedProducts(db *gorm.DB) error {
	return db.Exec(`
		INSERT INTO "Product" (created_at, updated_at, deleted_at, name, stock, price)
		SELECT NOW(), NOW(), NULL, name, stock, price
		FROM (VALUES
			('Óleo de Motor 5W-30 (1L)',        100, 3990),
			('Filtro de Óleo',                   80, 2490),
			('Filtro de Ar',                      60, 1990),
			('Pastilha de Freio Dianteira',       40, 8990),
			('Pastilha de Freio Traseira',        40, 7490),
			('Fluido de Freio DOT 4 (500ml)',     50, 1890),
			('Correia Dentada',                   20, 12990),
			('Vela de Ignição (unidade)',        120, 1490),
			('Fluido de Arrefecimento (1L)',      60, 2290),
			('Lâmpada H4 Halógena',             80, 1290)
		) AS t(name, stock, price)
		WHERE NOT EXISTS (SELECT 1 FROM "Product" WHERE deleted_at IS NULL);
	`).Error
}

func seedMaintenances(db *gorm.DB) error {
	return db.Exec(`
		INSERT INTO "Maintenance" (created_at, updated_at, deleted_at, name, price, category_id)
		SELECT NOW(), NOW(), NULL, name, price, category_id
		FROM (VALUES
			('Troca de Óleo e Filtro',               8990,  'padrao'),
			('Revisão de Freios',                    14990, 'padrao'),
			('Alinhamento e Balanceamento',           9990,  'padrao'),
			('Troca de Correia Dentada',             29990, 'utilitario'),
			('Revisão de Suspensão',                 24990, 'utilitario'),
			('Diagnóstico Eletrônico',               19990, 'utilitario'),
			('Revisão Completa (veículo comercial)', 59990, 'comercial'),
			('Troca de Embreagem',                   49990, 'comercial'),
			('Revisão de Injeção Eletrônica',        34990, 'comercial'),
			('Revisão Premium Completa',             99990, 'premium'),
			('Polimento e Cristalização',            39990, 'premium'),
			('Higienização de Ar-Condicionado',      14990, 'premium')
		) AS t(name, price, category_id)
		WHERE NOT EXISTS (SELECT 1 FROM "Maintenance" WHERE deleted_at IS NULL);
	`).Error
}
