package repository

import (
	"context"
	"database/sql"
	"example.com/pki-control-plane/internal/pki/domain"
)

type PostgresStore struct{ db *sql.DB }

func NewPostgresStore(db *sql.DB) *PostgresStore                                    { return &PostgresStore{db: db} }
func (p *PostgresStore) Ping(ctx context.Context) error                             { return p.db.PingContext(ctx) }
func (p *PostgresStore) SaveCA(context.Context, *domain.CertificateAuthority) error { return nil }
func (p *PostgresStore) GetCA(context.Context, domain.ID) (domain.CertificateAuthority, error) {
	return domain.CertificateAuthority{}, sql.ErrNoRows
}
func (p *PostgresStore) ListCAs(context.Context) []domain.CertificateAuthority         { return nil }
func (p *PostgresStore) SaveProfile(context.Context, *domain.CertificateProfile) error { return nil }
func (p *PostgresStore) GetProfile(context.Context, domain.ID) (domain.CertificateProfile, error) {
	return domain.CertificateProfile{}, sql.ErrNoRows
}
func (p *PostgresStore) SaveCertificate(context.Context, *domain.Certificate) error { return nil }
func (p *PostgresStore) GetCertificate(context.Context, domain.ID) (domain.Certificate, error) {
	return domain.Certificate{}, sql.ErrNoRows
}
func (p *PostgresStore) FindCertificateByIdempotency(context.Context, string, domain.Tenant) (domain.Certificate, bool) {
	return domain.Certificate{}, false
}
func (p *PostgresStore) ListCertificates(context.Context) []domain.Certificate { return nil }
func (p *PostgresStore) SaveAudit(context.Context, domain.AuditRecord) error   { return nil }
func (p *PostgresStore) Audits(context.Context) []domain.AuditRecord           { return nil }
func (p *PostgresStore) SaveOutbox(context.Context, domain.OutboxEvent) error  { return nil }
func (p *PostgresStore) Outbox(context.Context) []domain.OutboxEvent           { return nil }
