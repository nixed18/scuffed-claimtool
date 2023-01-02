Uses a couple APIs to tell you once every 10 seconds what the current highest TX is that will claim the next block.

Does NOT detect if someone has a valid claim transaction with a high fee, who also includes a secondary transaction that would come adter the first. It detects this as said address appearing in over 1 tx, which would normally disqualify it from claiming.

Run using -limit=X to change how many transactions deep into the mempool it searches for a valid claim transaction. If a TX cannot be found, it'll print out how deep it went, and the lowest fee it got to. Larger numbers take longer to pull, default is limit=100