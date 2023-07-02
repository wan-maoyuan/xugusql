package drive

import (
	"errors"
)

type xugusqlTx struct {
	tconn *xugusqlConn
}

func (self *xugusqlTx) Commit() error {
	if self.tconn == nil {
		return errors.New("Invalid connection")
	}
	err := self.tconn.exec("commit;")
	if err != nil {
		return err
	}

	return self.tconn.exec("set auto_commit on;")
}

func (self *xugusqlTx) Rollback() error {

	if self.tconn == nil {
		return errors.New("Invalid connection")
	}
	err := self.tconn.exec("rollback;")
	if err != nil {
		return err
	}

	return self.tconn.exec("set auto_commit on;")

}
