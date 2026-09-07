package pkg

import (
    "errors"
    "fmt"
    "net"

	"gorm.io/gorm"
    
    "github.com/mattn/go-sqlite3"
)

// SQLite error codes / extended codes
// https://www.sqlite.org/rescode.html
const (
    SqliteOk           = 0
    SqliteError        = 1
    SqliteInternal     = 2
    SqlitePerm         = 3
    SqliteAbort        = 4
    SqliteBusy         = 5
    SqliteLocked       = 6
    SqliteNomem        = 7
    SqliteReadonly     = 8
    SqliteInterrupt    = 9
    SqliteIoerr        = 10
    SqliteCorrupt      = 11
    SqliteFull         = 13
    SqliteCantopen     = 14
    SqliteProtocol     = 15
    SqliteEmpty        = 16
    SqliteSchema       = 17
    SqliteToobig       = 18
    SqliteConstraint   = 19
    SqliteMismatch     = 20
    SqliteMisuse       = 21
    SqliteNofstype     = 22
    SqliteAuth         = 23
    SqliteFormat       = 24
    SqliteRange        = 25
    SqliteNotadb       = 26
    SqliteNotice       = 27
    SqliteWarning      = 28

    // Extended codes for constraints
    SqliteConstraintCheck       = SqliteConstraint | (1 << 8)
    SqliteConstraintCommithook  = SqliteConstraint | (2 << 8)
    SqliteConstraintForeignKey  = SqliteConstraint | (3 << 8)
    SqliteConstraintFunction    = SqliteConstraint | (4 << 8)
    SqliteConstraintNotNull     = SqliteConstraint | (5 << 8)
    SqliteConstraintPrimaryKey  = SqliteConstraint | (6 << 8)
    SqliteConstraintTrigger     = SqliteConstraint | (7 << 8)
    SqliteConstraintUnique      = SqliteConstraint | (8 << 8)
    SqliteConstraintVtab        = SqliteConstraint | (9 << 8)
    SqliteConstraintRowid       = SqliteConstraint | (10 << 8)
    SqliteConstraintDatatype    = SqliteConstraint | (11 << 8)

    // Extended codes for busy/locked
    SqliteBusyRecovery          = SqliteBusy | (1 << 8)
    SqliteBusySnapshot          = SqliteBusy | (2 << 8)
    SqliteLockedSharedcache     = SqliteLocked | (1 << 8)
    SqliteLockedVtab            = SqliteLocked | (2 << 8)
)

// DB infrastructure errors
var (
    ErrAlreadyExists         = errors.New("record already exists")
    ErrNotFound              = errors.New("record not found")
    ErrForeignKeyViolation   = errors.New("related record not found or still referenced")
    ErrNotNullViolation      = errors.New("required field is missing")
    ErrCheckViolation        = errors.New("value does not satisfy constraint")
    ErrDataTooLong           = errors.New("data exceeds maximum length")
    ErrInvalidDataFormat     = errors.New("invalid data format")
    ErrDeadlock              = errors.New("deadlock detected, please retry")
    ErrSerializationFailure  = errors.New("transaction conflict, please retry")
    ErrConnectionFailed      = errors.New("database connection failed")
    ErrTooManyConnections    = errors.New("database is overloaded, too many connections")
    ErrQueryCanceled         = errors.New("query was canceled due to timeout")
    ErrInsufficientPrivilege = errors.New("insufficient database privilege")
    ErrUndefinedTable        = errors.New("table does not exist")
    ErrUndefinedColumn       = errors.New("column does not exist")
    ErrInternal              = errors.New("internal database error")
)

// HandleDBError maps driver and GORM errors to domain errors to prevent leakage.
func HandleDBError(err error) error {
    if err == nil {
        return nil
    }

    if errors.Is(err, gorm.ErrRecordNotFound) {
        return ErrNotFound
    }
    if errors.Is(err, gorm.ErrDuplicatedKey) {
        return ErrAlreadyExists
    }
    if errors.Is(err, gorm.ErrForeignKeyViolated) {
        return ErrForeignKeyViolation
    }
    if errors.Is(err, gorm.ErrCheckConstraintViolated) {
        return ErrCheckViolation
    }

    var sqliteErr sqlite3.Error
    if errors.As(err, &sqliteErr) {
        return handleSQLiteError(&sqliteErr)
    }

    var netErr *net.OpError
    if errors.As(err, &netErr) {
        return fmt.Errorf("%w: %s", ErrConnectionFailed, netErr.Op)
    }

    return fmt.Errorf("%w: %s", ErrInternal, err.Error())
}

func handleSQLiteError(err *sqlite3.Error) error {
    switch err.ExtendedCode {
    case SqliteConstraintUnique, SqliteConstraintPrimaryKey:
        return ErrAlreadyExists

    case SqliteConstraintForeignKey:
        return ErrForeignKeyViolation

    case SqliteConstraintNotNull:
        return ErrNotNullViolation

    case SqliteConstraintCheck:
        return ErrCheckViolation

    case SqliteBusy, SqliteBusyRecovery, SqliteBusySnapshot, SqliteLocked, SqliteLockedSharedcache, SqliteLockedVtab:
        return ErrDeadlock

    case SqliteFull:
        return fmt.Errorf("%w: disk full", ErrInternal)

    case SqliteNomem:
        return fmt.Errorf("%w: out of memory", ErrInternal)

    case SqliteInterrupt:
        return ErrQueryCanceled

    case SqlitePerm, SqliteReadonly, SqliteAuth:
        return ErrInsufficientPrivilege

    case SqliteToobig:
        return fmt.Errorf("%w: data exceeds maximum length", ErrDataTooLong)

    case SqliteMismatch, SqliteRange:
        return fmt.Errorf("%w: %s", ErrInvalidDataFormat, err.Error())

    case SqliteCorrupt, SqliteNotadb:
        return fmt.Errorf("%w: database corruption detected", ErrInternal)

    default:
        switch err.Code {
        case SqliteConstraint:
            return fmt.Errorf("%w: constraint violation", ErrCheckViolation)
        case SqliteBusy, SqliteLocked:
            return ErrDeadlock
        case SqliteFull:
            return fmt.Errorf("%w: disk full", ErrInternal)
        case SqliteNomem:
            return fmt.Errorf("%w: out of memory", ErrInternal)
        default:
            return fmt.Errorf("%w: sqlite_code=%d extended_code=%d msg=%s", ErrInternal, err.Code, err.ExtendedCode, err.Error())
        }
    }
}