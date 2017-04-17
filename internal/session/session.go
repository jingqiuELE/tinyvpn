//session includes the type definition for a client's session.
//A session has an index, which maps to a randomly generated session secret.
package session

const IndexLen = 6
const SecretLen = 32

type Index [IndexLen]byte
type Secret [SecretLen]byte
