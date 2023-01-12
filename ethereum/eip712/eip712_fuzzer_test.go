package eip712_test

import (
	"fmt"
	"math/rand"
	"strings"

	"github.com/evmos/ethermint/ethereum/eip712"
	"github.com/tidwall/gjson"
	"github.com/tidwall/sjson"
)

// Generates many random payloads with different JSON values to ensure
// that Flattening works across all inputs.
func (suite *EIP712TestSuite) TestRandomPayloadFlattening() {
	for i := 0; i < 10; i++ {
		suite.Run(fmt.Sprintf("Flatten%d", i), func() {
			payload := suite.generateRandomPayload(i)

			flattened, numMessages, err := eip712.FlattenPayloadMessages(payload)

			suite.Require().NoError(err)
			suite.Require().Equal(numMessages, i)

			suite.verifyPayloadAgainstFlattened(payload, flattened)
		})
	}
}

func (suite *EIP712TestSuite) generateRandomPayload(numMessages int) gjson.Result {
	payload := suite.createRandomJSONObject().Raw
	msgs := make([]gjson.Result, numMessages)

	for i := 0; i < numMessages; i++ {
		m := suite.createRandomJSONObject()
		msgs[i] = m
	}

	payload, err := sjson.Set(payload, "msgs", msgs)
	suite.Require().NoError(err)

	return gjson.Parse(payload)
}

func (suite *EIP712TestSuite) createRandomJSONObject() gjson.Result {
	var err error
	payloadRaw := ""

	numFields := suite.randomInRange(0, 16)
	for i := 0; i < numFields; i++ {
		key := suite.generateRandomString(12, 36)

		randField := suite.createRandomJSONField(i, 0)
		payloadRaw, err = sjson.Set(payloadRaw, key, randField)
		suite.Require().NoError(err)
	}

	return gjson.Parse(payloadRaw)
}

func (suite *EIP712TestSuite) createRandomJSONField(t int, depth int) interface{} {
	switch t % 5 {
	case 0:
		// Rand bool
		return rand.Intn(2) == 0
	case 1:
		// Rand string
		return suite.generateRandomString(10, 48)
	case 2:
		// Rand num
		return (rand.Float64() - 0.5) * 1000000
	case 3, 4:
		// Rand array (3) or object (4)
		arr := make([]interface{}, rand.Intn(10))
		obj := make(map[string]interface{})

		for i := range arr {
			fieldType := rand.Intn(5)
			if depth == 5 {
				// Max depth
				fieldType = rand.Intn(3)
			}

			randField := suite.createRandomJSONField(fieldType, depth+1)

			if t%5 == 3 {
				arr[i] = randField
			} else {
				obj[suite.generateRandomString(10, 48)] = randField
			}
		}

		if t%5 == 3 {
			return arr
		}
		return obj
	default:
		// Null
		return nil
	}
}

func (suite *EIP712TestSuite) generateRandomString(minLength int, maxLength int) string {
	bzLen := suite.randomInRange(minLength, maxLength)
	bz := make([]byte, bzLen)

	for i := 0; i < bzLen; i++ {
		bz[i] = byte(suite.randomInRange(65, 127))
	}

	str := string(bz[:])
	// Remove control characters
	str = strings.ReplaceAll(str, "{", "")
	str = strings.ReplaceAll(str, "}", "")
	str = strings.ReplaceAll(str, "]", "")
	str = strings.ReplaceAll(str, "[", "")

	return str
}

func (suite *EIP712TestSuite) randomInRange(min int, max int) int {
	return rand.Intn(max-min) + min
}
