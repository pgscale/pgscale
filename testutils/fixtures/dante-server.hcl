dante {
  bind_addr = "127.0.0.1"
  bind_port = 6957

  auth {
    users = {
      admin = {
        auth_type = "md5"
        hash = "558e292c17f2b28142ab3a85d92952fd"
      }
      dbuser = {
        auth_type = "password"
        password = "1234"
      }
    }
  }

  logging {
    verbosity = 6
    level = "DEBUG"
    output = "stderr"
    perm = 0644
  }

  postgresql {
    database "postgres" {
      parameters = {
        user = "postgres"
        host = "localhost"
        port = 5432
      }

      log_statements = true
      reset_query = "DISCARD ALL"

      connection_pool {
        policy              = "session"
        max_conns           = 50
        min_conns           = 0
        max_conn_lifetime   = "1h"
        max_conn_idle_time  = "15m"
        health_check_period = "1m"
      }
    }

    database "somedatabase" {
      parameters = {
        user = "postgres"
        host = "localhost"
        port = 5432
      }

      connection_pool {
        policy              = "statement"
        max_conns           = 50
        min_conns           = 0
        max_conn_lifetime   = "1h"
        max_conn_idle_time  = "15m"
        health_check_period = "1m"
      }

      log_statements = true
      reset_query = "DISCARD ALL"

      cache "public" {
        table "profile" {
          max_idle_duration = "60m"
          ttl_duration      = "10m"
          max_keys          = 500000
          lru_samples       = 20
          eviction_policy   = "NONE"
          storage_engine    = "kvstore"
        }

        table "users" {
          max_idle_duration = "60m"
          ttl_duration      = "10m"
          max_keys          = 500000
          lru_samples       = 20
          eviction_policy   = "NONE"
          storage_engine    = "kvstore"
        }
      }

      cache "different-schema" {
        table "users" {
          max_idle_duration = "60s"
          ttl_duration      = "10s"
          max_keys          = 500000
          lru_samples       = 20
          eviction_policy   = "NONE"
          storage_engine    = "kvstore"
        }
      }

    }
  }
}

olric {
  bind_addr = "0.0.0.0"
  bind_port = 3320
  bootstrap_timeout = "5s"
  keepalive_period = "300s"
  partition_count = 13
  read_quorum = 1
  write_quorum = 1
  member_count_quorum = 1
  read_repair = false
  replica_count = 1
  replication_mode = 0
  routing_table_push_interval = "1m"
  serializer = "msgpack"

  client {
    dial_timeout = "-1s"
    keep_alive = "15s"
    max_conn = 100
    min_conn = 1
    read_timeout = "3s"
    write_timeout = "3s"
  }

  logging {
    level = "DEBUG"
    output = "stderr"
    verbosity = 6
  }

  memberlist {
    bind_addr = "0.0.0.0"
    bind_port = 3322
    enable_compression = false
    environment = "local"
    join_retry_interval = "1ms"
    max_join_attempts = 1
    peers = [
    ]
  }

  storage_engines {
    engines = {
      kvstore = {
        table_size: 1048576
      }
    }
  }
}

