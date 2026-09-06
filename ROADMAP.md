# Fork roadmap

The fork aims to make Russian morphological analysis and inflection in Go easier to embed,
with independent analyzer instances and explicit compatibility expectations.
This is an experimental fork, not a promise of full parity with a Python release.

## Completed

- Added `NewMorphAnalyzer` for independent instances.
- Cached `GetMorphInstance` analyzers by dictionary path.
- Removed hidden dependencies on a global analyzer from internal analyzers.
- Kept the owning analyzer on each parse result for inflection and lexeme operations.
- Protected the grammeme cache and returned copies to callers.
- Added lifecycle and concurrent inflection tests.

## Next steps

1. **Define compatibility.** Build a reproducible comparison corpus for Python morphology
   libraries. Record dictionary-version differences separately from algorithm differences;
   do not claim equivalence to either pymorphy2 or pymorphy3 before this comparison exists.
2. **Expand concurrency coverage.** Exercise shared and independent analyzers under
   `go test -race ./...`, including parsing, inflection, and dictionary loading failures.
3. **Run checks in CI.** Test the declared minimum Go version and the current stable version,
   including lifecycle tests, the race detector, and `go vet`.
4. **Make dictionary setup reproducible.** Specify the dictionary version and update procedure,
   preserve source/data licensing, and provide a tested setup that avoids manual copying.
5. **Add integration examples.** Show parsing, inflection, and agreement with numbers in
   standalone programs with documented inputs and outputs.
6. **Prepare a release.** Document the public API, compatibility results, and limitations,
   then publish a version after the checks above pass.
