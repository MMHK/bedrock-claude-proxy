/**
 * @type {import('jest').Config}
 */
const config = {
    testEnvironment: "node",
    setupFiles: ['<rootDir>/jest.setup.js'],
    transform: {
        "^.+\\.js$": "babel-jest"
    },
    extensionsToTreatAsEsm: ['.ts', '.tsx', '.jsx'],
    "testMatch": [
        "./**/*.test.js"
    ],
}

module.exports = config;