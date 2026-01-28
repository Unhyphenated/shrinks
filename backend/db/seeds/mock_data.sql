-- backend/db/seeds/mock_data.sql
-- Run with: psql $DATABASE_URL -f backend/db/seeds/mock_data.sql

-- =============================================
-- 1. Create test users
-- =============================================
-- Password for all: "password123" (bcrypt hash)
INSERT INTO users (email, password_hash) VALUES
  ('demo@shrinks.io', '$2a$10$N9qo8uLOickgx2ZMRZoMy.MqrqJLlBX5qUhWKLzITn1bVBb3gPKXK'),
  ('test@example.com', '$2a$10$N9qo8uLOickgx2ZMRZoMy.MqrqJLlBX5qUhWKLzITn1bVBb3gPKXK'),
  ('alice@example.com', '$2a$10$N9qo8uLOickgx2ZMRZoMy.MqrqJLlBX5qUhWKLzITn1bVBb3gPKXK')
ON CONFLICT (email) DO NOTHING;

-- =============================================
-- 2. Create links for demo user
-- =============================================
DO $$
DECLARE
  demo_user_id BIGINT;
  link_id BIGINT;
  i INT;
  urls TEXT[] := ARRAY[
    'https://github.com/trending',
    'https://news.ycombinator.com',
    'https://stackoverflow.com/questions',
    'https://reddit.com/r/programming',
    'https://dev.to',
    'https://medium.com/tag/technology',
    'https://twitter.com',
    'https://linkedin.com/feed',
    'https://youtube.com/watch?v=dQw4w9WgXcQ',
    'https://google.com/search?q=golang+best+practices',
    'https://docs.docker.com/get-started',
    'https://kubernetes.io/docs/home',
    'https://aws.amazon.com/ec2',
    'https://cloud.google.com/run',
    'https://vercel.com/docs',
    'https://nextjs.org/learn',
    'https://react.dev/learn',
    'https://go.dev/doc/effective_go',
    'https://redis.io/docs',
    'https://postgresql.org/docs'
  ];
  domains TEXT[] := ARRAY['Desktop', 'Mobile', 'Tablet'];
  browsers TEXT[] := ARRAY['Chrome 120', 'Safari 17', 'Firefox 121', 'Edge 120'];
  os_list TEXT[] := ARRAY['Windows', 'macOS', 'iOS', 'Android', 'Linux'];
BEGIN
  -- Get demo user ID
  SELECT id INTO demo_user_id FROM users WHERE email = 'demo@shrinks.io';

  IF demo_user_id IS NULL THEN
    RAISE NOTICE 'Demo user not found, skipping...';
    RETURN;
  END IF;

  -- Create 20 links
  FOR i IN 1..20 LOOP
    INSERT INTO links (id, user_id, short_code, long_url, created_at)
    VALUES (
      100000 + i,
      demo_user_id,
      encode(sha256(random()::text::bytea), 'hex')::varchar(6),
      urls[i],
      NOW() - (random() * interval '30 days')
    )
    ON CONFLICT DO NOTHING
    RETURNING id INTO link_id;

    -- Skip if link wasn't created
    IF link_id IS NULL THEN
      SELECT id INTO link_id FROM links WHERE long_url = urls[i] AND user_id = demo_user_id LIMIT 1;
    END IF;

    -- Add 10-100 analytics events per link
    FOR j IN 1..(10 + floor(random() * 90)::int) LOOP
      INSERT INTO analytics (link_id, ip_address, user_agent, device_type, browser, os, clicked_at)
      VALUES (
        link_id,
        (floor(random() * 255)::int || '.' || floor(random() * 255)::int || '.' || floor(random() * 255)::int || '.0')::inet,
        'Mozilla/5.0 Mock User Agent',
        domains[1 + floor(random() * 3)::int],
        browsers[1 + floor(random() * 4)::int],
        os_list[1 + floor(random() * 5)::int],
        NOW() - (random() * interval '30 days')
      );
    END LOOP;
  END LOOP;

  RAISE NOTICE 'Created 20 links with analytics for demo user';
END $$;

-- =============================================
-- 3. Create some anonymous links (no user)
-- =============================================
INSERT INTO links (id, user_id, short_code, long_url, created_at)
SELECT
  100100 + generate_series,
  NULL,
  encode(sha256(('anon' || generate_series)::bytea), 'hex')::varchar(6),
  'https://example.com/page/' || generate_series,
  NOW() - (random() * interval '60 days')
FROM generate_series(1, 50)
ON CONFLICT DO NOTHING;

-- =============================================
-- 4. Summary
-- =============================================
SELECT 'Users' as table_name, COUNT(*) as count FROM users
UNION ALL
SELECT 'Links', COUNT(*) FROM links
UNION ALL
SELECT 'Analytics Events', COUNT(*) FROM analytics;