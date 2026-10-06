/*
antennaApp is a package for creating and curating blog, link blogs and social websites
Copyright (C) 2025 R. S. Doiel

This program is free software: you can redistribute it and/or modify
it under the terms of the GNU Affero General Public License as published by
the Free Software Foundation, either version 3 of the License, or
(at your option) any later version.

This program is distributed in the hope that it will be useful,
but WITHOUT ANY WARRANTY; without even the implied warranty of
MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
GNU Affero General Public License for more details.

You should have received a copy of the GNU Affero General Public License
along with this program.  If not, see <https://www.gnu.org/licenses/>.
*/
package antennaApp

import (
	"archive/zip"
	"encoding/json"
	"encoding/xml"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"net"
	"net/url"
	"os"
	"syscall"

	// 3rd Party Packages
	sqlite "github.com/glebarez/go-sqlite"
	"gopkg.in/yaml.v3"
)

/** ExitClass is one row of the workspace exit-code table (DR-0003): the class
 * name and the process exit code it maps to.
 *
 * Fields:
 *   Name (string) — the class name, e.g. "no_input".
 *   Code (int)    — the exit code, e.g. 66.
 *
 * Example:
 *   c := ExitCodeFor(err)
 *   fmt.Fprintf(os.Stderr, "%s (%d)\n", c.Name, c.Code)
 */
type ExitClass struct {
	Name string
	Code int
}

// The classes of the workspace table (DR-0003). 64 is deliberately absent:
// usage is 2.
var (
	classOK           = ExitClass{"ok", 0}
	classNegative     = ExitClass{"negative", 1}
	classUsage        = ExitClass{"usage", 2}
	classData         = ExitClass{"data", 65}
	classNoInput      = ExitClass{"no_input", 66}
	classUnavailable  = ExitClass{"unavailable", 69}
	classInternal     = ExitClass{"internal", 70}
	classCantCreate   = ExitClass{"cant_create", 73}
	classIO           = ExitClass{"io", 74}
	classTempFail     = ExitClass{"temp_fail", 75}
	classNoPermission = ExitClass{"no_permission", 77}
	classConfig       = ExitClass{"config", 78}
)

// exitClass is the unexported spelling used inside the package.
type exitClass = ExitClass

// allExitClasses lists every class so a test can check the table as a whole.
var allExitClasses = []exitClass{
	classOK, classNegative, classUsage, classData, classNoInput, classUnavailable,
	classInternal, classCantCreate, classIO, classTempFail, classNoPermission, classConfig,
}

// classedError is an error that carries its exit class. classify finds it with
// errors.As, so a classed error keeps its class through any amount of %w
// wrapping.
type classedError struct {
	class exitClass
	err   error
}

func (c *classedError) Error() string { return c.err.Error() }
func (c *classedError) Unwrap() error { return c.err }

// classedAs marks an existing error with a class and leaves nil alone.
func classedAs(class exitClass, err error) error {
	if err == nil {
		return nil
	}
	return &classedError{class: class, err: err}
}

// classErrorf is fmt.Errorf that marks its result with a class. %w works as it
// does in fmt.Errorf, so a wrapped cause stays reachable.
func classErrorf(class exitClass, format string, a ...any) error {
	return &classedError{class: class, err: fmt.Errorf(format, a...)}
}

// usageErrorf marks a mistake in the command line itself: a missing or surplus
// argument, an unknown verb, flag or help topic, or a bad value passed on the
// command line. Nothing was attempted.
func usageErrorf(format string, a ...any) error { return classErrorf(classUsage, format, a...) }

// negativef marks a command that ran correctly and whose answer is no: no such
// item, nothing found by a query, an operation the current state forbids.
func negativef(format string, a ...any) error { return classErrorf(classNegative, format, a...) }

// dataErrorf marks content the tool read that is wrong: a malformed front
// matter block, a document or database of the wrong kind.
func dataErrorf(format string, a ...any) error { return classErrorf(classData, format, a...) }

// noInputf marks a named input or the workspace itself as missing or unreadable.
func noInputf(format string, a ...any) error { return classErrorf(classNoInput, format, a...) }

// unavailablef marks a service the command needs as unreachable.
func unavailablef(format string, a ...any) error { return classErrorf(classUnavailable, format, a...) }

// internalErrorf marks a bug: a condition that should be impossible.
func internalErrorf(format string, a ...any) error { return classErrorf(classInternal, format, a...) }

// cantCreatef marks an output that cannot be created: the target exists or its
// directory cannot be made.
func cantCreatef(format string, a ...any) error { return classErrorf(classCantCreate, format, a...) }

// ioErrorf marks a read or write that failed part way.
func ioErrorf(format string, a ...any) error { return classErrorf(classIO, format, a...) }

// configErrorf marks a configuration file that is present but wrong.
func configErrorf(format string, a ...any) error { return classErrorf(classConfig, format, a...) }

// classify returns the exit class of err. The second result is false when
// nothing classified err, in which case the class returned is internal and
// ExitCodeFor reports it as 70. A class set explicitly with classedAs or
// classErrorf wins over everything the standard library would say.
func classify(err error) (exitClass, bool) {
	if err == nil {
		return classOK, true
	}
	var ce *classedError
	if errors.As(err, &ce) {
		return ce.class, true
	}
	var sqlErr *sqlite.Error
	if errors.As(err, &sqlErr) {
		return sqliteClass(sqlErr.Code()), true
	}
	if code, ok := mattnSQLiteCode(err); ok {
		return sqliteClass(code), true
	}
	// Content that would not decode is wrong content, whichever verb read it.
	var syn *json.SyntaxError
	var typ *json.UnmarshalTypeError
	var yTyp *yaml.TypeError
	var xmlSyn *xml.SyntaxError
	if errors.As(err, &syn) || errors.As(err, &typ) || errors.As(err, &yTyp) ||
		errors.As(err, &xmlSyn) || errors.Is(err, zip.ErrFormat) ||
		errors.Is(err, io.ErrUnexpectedEOF) {
		return classData, true
	}
	switch {
	case errors.Is(err, fs.ErrNotExist):
		return classNoInput, true
	case errors.Is(err, fs.ErrPermission):
		return classNoPermission, true
	case errors.Is(err, fs.ErrExist):
		return classCantCreate, true
	}
	// A port already in use is a busy resource, not an unreachable service:
	// a retry may succeed. Checked before the network types below, which would
	// call it unavailable.
	if errors.Is(err, syscall.EADDRINUSE) {
		return classTempFail, true
	}
	// Network errors, by concrete type first. A refused connection wraps an
	// *os.SyscallError, so the file errors below would call it a failed write;
	// but the net.Error interface cannot be tested first either, because a bare
	// syscall.Errno (what a *fs.PathError wraps) satisfies it too.
	var oe *net.OpError
	var ue *url.Error
	var de *net.DNSError
	if errors.As(err, &oe) || errors.As(err, &ue) || errors.As(err, &de) {
		return classUnavailable, true
	}
	var pe *fs.PathError
	var le *os.LinkError
	var se *os.SyscallError
	if errors.As(err, &pe) || errors.As(err, &le) || errors.As(err, &se) {
		return classIO, true
	}
	var ne net.Error
	if errors.As(err, &ne) {
		return classUnavailable, true
	}
	return classInternal, false
}

/** ExitCodeFor returns the exit class a command that failed with err should
 * exit with, following the workspace convention (DR-0003). nil is ok (0). An
 * error nothing classified is internal (70), never negative (1), so a site
 * nobody classified shows up as an internal error instead of hiding as a
 * negative answer.
 *
 * Parameters:
 *   err (error) — the error a command returned, or nil.
 *
 * Returns:
 *   ExitClass — the class name and exit code.
 *
 * Example:
 *   if err := app.Run(in, out, eout, cfg, action, args); err != nil {
 *       fmt.Fprintln(eout, err)
 *       os.Exit(antennaApp.ExitCodeFor(err).Code)
 *   }
 */
func ExitCodeFor(err error) ExitClass {
	class, _ := classify(err)
	return class
}

// sqliteClass maps a SQLite result code (the low byte of an extended code) to
// an exit class. Both drivers this package links report the same codes.
func sqliteClass(code int) exitClass {
	switch code & 0xff {
	case 5, 6: // SQLITE_BUSY, SQLITE_LOCKED
		return classTempFail
	case 11, 26: // SQLITE_CORRUPT, SQLITE_NOTADB
		return classData
	case 19: // SQLITE_CONSTRAINT, whichever extended code
		return classData
	}
	return classIO
}

// asCreate marks a failure to create an output as cant_create (73) when the
// operating system refused to open, make or create it, and leaves everything
// else as it was: a refused permission stays no_permission (77), a write that
// failed part way stays io (74), and an error that already has a class keeps
// it. Wrap the error from os.Create, os.WriteFile, os.MkdirAll and friends at
// the site that creates an output.
func asCreate(err error) error {
	var ce *classedError
	if err == nil || errors.As(err, &ce) || errors.Is(err, fs.ErrPermission) {
		return err
	}
	var pe *fs.PathError
	if errors.Is(err, fs.ErrExist) ||
		(errors.As(err, &pe) && (pe.Op == "open" || pe.Op == "mkdir" || pe.Op == "create")) {
		return classedAs(classCantCreate, err)
	}
	return err
}

// failureTally counts the outcomes of a bulk command, which does not stop at
// the first failure: it processes everything, then exits with the class of the
// first failure and says how many failed (DR-0003 item 4). Items are
// processed in a fixed order (configuration order), so "first" does not depend
// on directory or map order.
type failureTally struct {
	total  int
	failed int
	first  error
}

// add records the outcome of one item; a nil err counts it as a success.
func (f *failureTally) add(err error) {
	f.total++
	if err != nil {
		f.failed++
		if f.first == nil {
			f.first = err
		}
	}
}

// err returns nil when nothing failed. Otherwise it returns an error that
// carries the first failure's class and names the counts, e.g.
// "1 of 2 feeds failed: <first failure>". what is the plural noun counted.
func (f *failureTally) err(what string) error {
	if f.failed == 0 {
		return nil
	}
	return fmt.Errorf("%d of %d %s failed: %w", f.failed, f.total, what, f.first)
}
