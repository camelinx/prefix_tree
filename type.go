package prefix_tree

type OpResult int

const (
    Err   OpResult = iota
    Ok
    Dup
    Match
    PartialMatch
    NoMatch
)

type MatchType int

const (
    Exact   MatchType = iota
    Partial
)

type ReadLockFn func( interface{ } )( )
type ReadUnlockFn func( interface{ } )( )
type WriteLockFn func( interface{ } )( )
type UnlockFn func( interface{ } )( )

type AddrTree interface {
    SetLockHandlers( interface{ }, ReadLockFn, ReadUnlockFn, WriteLockFn, UnlockFn )( )
    Insert( string, interface{ } )( OpResult, error )
    Delete( string )( OpResult, interface{ }, error )
    Search( string )( OpResult, interface{ }, error )
    SearchExact( string )( OpResult, interface{ }, error )
    GetNodesCount( )( uint64 )
}
