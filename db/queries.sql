-- Part 2: Data Handling & Queries

-- 1) Total amount of completed transactions per user (sender)
SELECT
  sender AS user,
  SUM(amount) AS total_completed_amount
FROM transactions
WHERE status = 'completed'
GROUP BY sender
ORDER BY total_completed_amount DESC;

-- 2) Top 5 users by transaction volume (sum of amounts) in the last 30 days
SELECT
  sender AS user,
  SUM(amount) AS volume_last_30d
FROM transactions
WHERE status = 'completed'
  AND created_at >= NOW() - INTERVAL 30 DAY
GROUP BY sender
ORDER BY volume_last_30d DESC
LIMIT 5;

-- 3) All users with more than 3 failed transactions (sender)
SELECT
  sender AS user,
  COUNT(*) AS failed_count
FROM transactions
WHERE status = 'failed'
GROUP BY sender
HAVING COUNT(*) > 3
ORDER BY failed_count DESC;


