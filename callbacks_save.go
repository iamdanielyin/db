package db

func beginTransactionCallback(s *Scope) {
	tx, err := s.Session.StartTransaction()
	if err != nil {
		s.AddError(err).Skip()
		return
	}
	s.Store().Store("db:tx", tx)
}

func commitOrRollbackTransactionCallback(s *Scope) {
	if v, has := s.Store().Load("db:tx"); has {
		tx := v.(Tx)
		if s.HasError() {
			s.AddError(tx.Rollback())
			return
		}
		s.AddError(tx.Commit())
	}
}
