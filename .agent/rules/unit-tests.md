---
trigger: glob
globs: *_test.go
---

Unit tests

## Planning Rules

1. Identify Items to Test:

  - just show the code, don't need to explain anything.
  - Determine if a test suite (for structs with dependencies) or individual test functions (for standalone functions) should be used.
  - Identify dependencies or side effects that may require mocks or stubs.
  - Consider error conditions and how to test them.

2. Defining Test Cases:

  - Number each test case clearly (e.g., TestFunction_ValidInput_ReturnsExpectedResult, TestFunction_EmptyInput_ReturnsError).
  - Each test case should concisely describe the scenario being tested.

## Implementation Rules

### CRITICAL: Package Naming
- **ALWAYS** use the `_test` suffix for the package name in test files
- **NEVER** duplicate the package declaration
- Example of CORRECT package declaration:
```go
  package mypackage_test
```
- Example of INCORRECT package declaration:
```go
  package mypackage
  package mypackage  // ❌ NEVER duplicate
```

1. Test Suites for Structs with Dependencies:

  - Use suite.Suite from testify.
  - Create a suite struct that contains a sut (System Under Test) field.
  - Implement a SetupTest method to initialize the sut and its dependencies.
  - Do not use .AssertExpectations(s.T()).
  - Identify the constructor for the type and use it. Generally the name is NewTypeName, example NewAccountValidator(deps...)
  - **Package name must ALWAYS use `_test` suffix** (e.g., `package mypackage_test`)
  - For test suite always use suite instead of assert, example: 
    - `suite.Equal(v, 10)`
  - For error assertion on the suite tests, always use `suite.Require()`: 
    - `suite.Require().ErrorIs(err, expectedErr)`
    - `suite.Require().Error(err)` 
  - For error assertion on the non suite tests, always use require: 
    - `require.ErrorIs(t, err, expectedErr)`
    - `require.Error(t, err)`

- Test suite example:
```go
package mypackage_test

import (
	"testing"
	"github.com/stretchr/testify/suite"
)

type MyStructTestSuite struct {
	suite.Suite
	sut *mypackage.MyStruct
}

func (s *MyStructTestSuite) SetupTest() {
	s.sut = mypackage.NewMyStruct()
}

func TestMyStructSuite(t *testing.T) {
	suite.Run(t, new(MyStructTestSuite))
}

func (s *MyStructTestSuite) TestSomeMethod() {
	// Arrange
	
	// Act
	
	// Assert
}
```

1.1 On mocks always pass `mock.Anything` for contexts.
Example: 
```go
s.recurringPaymentRepoMock.On("Update", mock.Anything, payment).Return(nil)
```

2. Tests for Functions Without Instances:
  
  - Create individual test functions using `func TestXxx(t *testing.T)`.
  - Use `t.Run` for subtests covering different scenarios.
  - **Package name must ALWAYS use `_test` suffix** (e.g., `package mypackage_test`)
  - Example:
```go
package mypackage_test

import "testing"

func TestSomeFunction(t *testing.T) {
	t.Run("scenario description", func(t *testing.T) {
		// Arrange
		
		// Act
		
		// Assert
	})
}
```

3. Test Coverage:
  
  - Include happy path scenarios.
  - Include edge cases.
  - Minimum test scenarios possible.
  - Include error handling.
  - Aim for at least 80% coverage.

4. Using Assertions:

  - Use testify assertion functions (e.g., `s.Equal(expected, actual)` instead of `assert.Equal(s.T(), expected, actual)`).
  - Avoid `.AssertExpectations(s.T())`.

5. Mocks and Stubs:

  - Use mocks or stubs to isolate dependencies as needed.
  - Mocks are already in place; no need to generate them.
  - All mocks are located in the tests/mocks directory.
  - Example imports:
     - `"auction/tests/mocks"`

6. Mock Naming Convention:

  - Mocks follow the pattern MockType, for example:
    - `mocks.MockUserRepository`
    - `mocks.MockTokenService`

7. Arrange-Act-Assert Pattern:

  - Each test should follow the AAA pattern:
    - Arrange: prepare environment and data
    - Act: perform the action being tested
    - Assert: verify the results
  - In the code the comments must be in the pattern below:
```go
    // Arrange
    
    // Act
    
    // Assert
```

8. Naming and Clarity:

  - Test names must clearly indicate what is being tested.
  - Add comments for complex test setups or assertions if needed.
  - Never use inline struct construction, always create a new variable and assigns the instance to it.
  - Each line has to have max 120 characters.

9. Never explain the test after generating the code. Just say "Tests Done, Oh Yeah!" when the process is finished.