package persistence

import (
 "context"
 "errors"
 "github.com/jackc/pgx/v5"
 "github.com/jackc/pgx/v5/pgxpool"
 "github.com/fikrimastor/mytax/services/api/internal/security"
)
type TenantDB struct{ pool *pgxpool.Pool; issuer,audience string }
func NewTenantDB(pool *pgxpool.Pool,issuer,audience string)(*TenantDB,error){
 if pool==nil||issuer==""||audience==""{return nil,errors.New("missing verified DB dependencies")}
 return &TenantDB{pool,issuer,audience},nil
}
// WithTenant starts a read-only transaction. No SQL runs before subject validation.
// Membership is checked in the same transaction before the callback runs.
func (db *TenantDB) WithTenant(ctx context.Context,id security.Identity,tenantID string,fn func(pgx.Tx)error)error{
 if err:=security.RequireTenant(id,db.issuer,db.audience,tenantID);err!=nil{return err}
 if fn==nil{return errors.New("callback required")}
 tx,err:=db.pool.BeginTx(ctx,pgx.TxOptions{AccessMode:pgx.ReadOnly})
 if err!=nil{return err}
 defer tx.Rollback(ctx)
 if _,err=tx.Exec(ctx,"select set_config('app.subject', $1, true)",id.Subject);err!=nil{return err}
 var member bool
 if err=tx.QueryRow(ctx,"select exists(select 1 from tenant_memberships where tenant_id=$1::uuid and subject_id=$2)",tenantID,id.Subject).Scan(&member);err!=nil{return err}
 if !member{return errors.New("tenant membership denied")}
 if err=fn(tx);err!=nil{return err}
 return tx.Commit(ctx)
}
func (db *TenantDB) CountEvidence(ctx context.Context,id security.Identity,tenantID string)(int64,error){
 var count int64
 err:=db.WithTenant(ctx,id,tenantID,func(tx pgx.Tx)error{
  return tx.QueryRow(ctx,"select count(*) from evidence_metadata where tenant_id=$1::uuid",tenantID).Scan(&count)
 })
 return count,err
}
