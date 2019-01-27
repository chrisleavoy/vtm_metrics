package cmd

import (
	"fmt"
	"github.com/chrisleavoy/vtm_metrics/3.10"
	"github.com/spf13/cobra"
	"log"
	"os"
	"strings"
	"sync"
)

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

var (
	wg sync.WaitGroup
)

var rootCmd = &cobra.Command{
	Use:   "vtm_metrics",
	Short: "SecurePulse VTM Metrics",
	Long:  "Fetches various stats from the VTM and outputs the given metrics in InfluxDB line protocol format",
	Run: func(cmd *cobra.Command, args []string) {
		// global flags
		url, _ := cmd.Flags().GetString(fURL)
		user, _ := cmd.Flags().GetString(fUsername)
		pass, _ := cmd.Flags().GetString(fPassword)
		verbose, _ := cmd.Flags().GetBool(fVerbose)
		skipSSLVerify, _ := cmd.Flags().GetBool(fSkipSSLVerify)

		v, reachable, err := vtm.NewVirtualTrafficManager(url, user, pass, !skipSSLVerify, verbose)

		if !reachable {
			log.Fatalf("ERROR %+v\n", err)
		}

		// there is no bulk-get for metrics from the VTM, so we need to send a number of HTTP GET.
		// We do them asynchronously to avoid delays.
		// vtm.VirtualTrafficManager.newConnector() was monkey patched to limit MaxConnsPerHost: 25
		doAsyncStatsGathering(v)

		wg.Wait()
	},
}

func doAsyncStatsGathering(v *vtm.VirtualTrafficManager) {
	wg.Add(1)
	go func() {
		defer wg.Done()
		globalStats(v)
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()
		vss, err := v.ListVirtualServers()
		if err != nil {
			log.Printf("ERROR: %+v\n", err)
		}
		if err == nil {
			wg.Add(len(*vss))
			for _, vs := range *vss {
				go func(vs string) {
					defer wg.Done()
					vsStats(v, vs)
				}(vs)
			}
		}
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()
		slms, err := v.ListServiceLevelMonitors()
		if err != nil {
			log.Printf("ERROR: %+v\n", err)
		}
		if err == nil {
			wg.Add(len(*slms))
			for _, slm := range *slms {
				go func(slm string) {
					defer wg.Done()
					slmStats(v, slm)
				}(slm)
			}
		}
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()
		rates, err := v.ListRates()
		if err != nil {
			log.Printf("ERROR: %+v\n", err)
		}
		if err == nil {
			wg.Add(len(*rates))
			for _, rate := range *rates {
				go func(rate string) {
					defer wg.Done()
					rateStats(v, rate)
				}(rate)
			}
		}
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()
		pools, err := v.ListPools()
		if err != nil {
			log.Printf("ERROR: %+v\n", err)
		}
		if err == nil {
			wg.Add(len(*pools))
			for _, pool := range *pools {
				go func(pool string) {
					defer wg.Done()
					poolStats(v, pool)
				}(pool)
			}
		}
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()
		nodes, err := v.ListNodes()
		if err != nil {
			log.Printf("ERROR: %+v\n", err)
		}
		if err == nil {
			wg.Add(len(*nodes))
			for _, node := range *nodes {
				go func(node string) {
					defer wg.Done()
					nodeStats(v, node)
				}(node)
			}
		}
	}()
}

func globalStats(v *vtm.VirtualTrafficManager) {
	s, err := v.GetGlobalsStatistics()
	if err != nil {
		log.Printf("ERROR: %+v\n", err)
		return
	}

	// TODO host tag may not be the target
	fmt.Printf(
		"vtm %s=%di,%s=%di,%s=%di,%s=%di,%s=%di,%s=%di,%s=%di,%s=%di,%s=%di,%s=%di,%s=%di,%s=%di,%s=%di,%s=%di,%s=%di,%s=%di,%s=%di,%s=%di,%s=%di,%s=%di,%s=%di,%s=%di,%s=%di,%s=%di,%s=%di,%s=%di,%s=%di,%s=%di,%s=%di,%s=%di,%s=%di,%s=%di,%s=%di,%s=%di,%s=%di,%s=%di,%s=%di,%s=%di,%s=%di,%s=%di,%s=%di,%s=%di,%s=%di,%s=%di,%s=%di,%s=%di,%s=%di,%s=%di,%s=%di,%s=%di,%s=%di,%s=%di,%s=%di,%s=%di,%s=%di,%s=%di,%s=%di,%s=%di,%s=%di,%s=%di,%s=%di,%s=%di,%s=%di,%s=%di,%s=%di,%s=%di,%s=%di,%s=%di,%s=%di,%s=%di,%s=%di,%s=%di,%s=%di,%s=%di,%s=%di,%s=%di,%s=%di,%s=%di,%s=%di,%s=%di,%s=%di,%s=%di\n",
		"data_entries", *s.Statistics.DataEntries,
		"data_memory_usage", *s.Statistics.DataMemoryUsage,
		"events_seen", *s.Statistics.EventsSeen,
		"hourly_peak_bytes_in_per_second", *s.Statistics.HourlyPeakBytesInPerSecond,
		"hourly_peak_bytes_out_per_second", *s.Statistics.HourlyPeakBytesOutPerSecond,
		"hourly_peak_requests_per_second", *s.Statistics.HourlyPeakRequestsPerSecond,
		"hourly_peak_ssl_connections_per_second", *s.Statistics.HourlyPeakSslConnectionsPerSecond,
		"num_idle_connections", *s.Statistics.NumIdleConnections,
		"number_child_processes", *s.Statistics.NumberChildProcesses,
		"number_dnsa_cache_hits", *s.Statistics.NumberDnsaCacheHits,
		"number_dnsa_requests", *s.Statistics.NumberDnsaRequests,
		"number_dnsptr_cache_hits", *s.Statistics.NumberDnsptrCacheHits,
		"number_dnsptr_requests", *s.Statistics.NumberDnsptrRequests,
		"number_snmp_bad_requests", *s.Statistics.NumberSnmpBadRequests,
		"number_snmp_get_bulk_requests", *s.Statistics.NumberSnmpGetBulkRequests,
		"number_snmp_get_next_requests", *s.Statistics.NumberSnmpGetNextRequests,
		"number_snmp_get_requests", *s.Statistics.NumberSnmpGetRequests,
		"number_snmp_unauthorised_requests", *s.Statistics.NumberSnmpUnauthorisedRequests,
		"ssl_cipher_3des_decrypts", *s.Statistics.SslCipher3DesDecrypts,
		"ssl_cipher_3des_encrypts", *s.Statistics.SslCipher3DesEncrypts,
		"ssl_cipher_aes_decrypts", *s.Statistics.SslCipherAesDecrypts,
		"ssl_cipher_aes_encrypts", *s.Statistics.SslCipherAesEncrypts,
		"ssl_cipher_aes_gcm_decrypts", *s.Statistics.SslCipherAesGcmDecrypts,
		"ssl_cipher_aes_gcm_encrypts", *s.Statistics.SslCipherAesGcmEncrypts,
		"ssl_cipher_decrypts", *s.Statistics.SslCipherDecrypts,
		"ssl_cipher_des_decrypts", *s.Statistics.SslCipherDesDecrypts,
		"ssl_cipher_des_encrypts", *s.Statistics.SslCipherDesEncrypts,
		"ssl_cipher_dh_agreements", *s.Statistics.SslCipherDhAgreements,
		"ssl_cipher_dh_generates", *s.Statistics.SslCipherDhGenerates,
		"ssl_cipher_dsa_signs", *s.Statistics.SslCipherDsaSigns,
		"ssl_cipher_dsa_verifies", *s.Statistics.SslCipherDsaVerifies,
		"ssl_cipher_ecdh_agreements", *s.Statistics.SslCipherEcdhAgreements,
		"ssl_cipher_ecdh_generates", *s.Statistics.SslCipherEcdhGenerates,
		"ssl_cipher_ecdsa_signs", *s.Statistics.SslCipherEcdsaSigns,
		"ssl_cipher_ecdsa_verifies", *s.Statistics.SslCipherEcdsaVerifies,
		"ssl_cipher_encrypts", *s.Statistics.SslCipherEncrypts,
		"ssl_cipher_rc4_decrypts", *s.Statistics.SslCipherRc4Decrypts,
		"ssl_cipher_rc4_encrypts", *s.Statistics.SslCipherRc4Encrypts,
		"ssl_cipher_rsa_decrypts", *s.Statistics.SslCipherRsaDecrypts,
		"ssl_cipher_rsa_decrypts_external", *s.Statistics.SslCipherRsaDecryptsExternal,
		"ssl_cipher_rsa_encrypts", *s.Statistics.SslCipherRsaEncrypts,
		"ssl_cipher_rsa_encrypts_external", *s.Statistics.SslCipherRsaEncryptsExternal,
		"ssl_client_cert_expired", *s.Statistics.SslClientCertExpired,
		"ssl_client_cert_invalid", *s.Statistics.SslClientCertInvalid,
		"ssl_client_cert_not_sent", *s.Statistics.SslClientCertNotSent,
		"ssl_client_cert_revoked", *s.Statistics.SslClientCertRevoked,
		"ssl_connections", *s.Statistics.SslConnections,
		"ssl_handshake_sslv2", *s.Statistics.SslHandshakeSslv2,
		"ssl_handshake_sslv3", *s.Statistics.SslHandshakeSslv3,
		"ssl_handshake_t_l_sv1", *s.Statistics.SslHandshakeTLSv1,
		"ssl_handshake_t_l_sv11", *s.Statistics.SslHandshakeTLSv11,
		"ssl_handshake_t_l_sv12", *s.Statistics.SslHandshakeTLSv12,
		"ssl_session_id_disk_cache_hit", *s.Statistics.SslSessionIdDiskCacheHit,
		"ssl_session_id_disk_cache_miss", *s.Statistics.SslSessionIdDiskCacheMiss,
		"ssl_session_id_mem_cache_hit", *s.Statistics.SslSessionIdMemCacheHit,
		"ssl_session_id_mem_cache_miss", *s.Statistics.SslSessionIdMemCacheMiss,
		"sys_cpu_busy_percent", *s.Statistics.SysCpuBusyPercent,
		"sys_cpu_idle_percent", *s.Statistics.SysCpuIdlePercent,
		"sys_cpu_system_busy_percent", *s.Statistics.SysCpuSystemBusyPercent,
		"sys_cpu_user_busy_percent", *s.Statistics.SysCpuUserBusyPercent,
		"sys_fds_free", *s.Statistics.SysFdsFree,
		"sys_mem_buffered", *s.Statistics.SysMemBuffered,
		"sys_mem_free", *s.Statistics.SysMemFree,
		"sys_mem_in_use", *s.Statistics.SysMemInUse,
		"sys_mem_swap_total", *s.Statistics.SysMemSwapTotal,
		"sys_mem_swapped", *s.Statistics.SysMemSwapped,
		"sys_mem_total", *s.Statistics.SysMemTotal,
		"time_last_config_update", *s.Statistics.TimeLastConfigUpdate,
		"total_backend_server_errors", *s.Statistics.TotalBackendServerErrors,
		"total_bad_dns_packets", *s.Statistics.TotalBadDnsPackets,
		"total_bytes_in", *s.Statistics.TotalBytesIn,
		"total_bytes_in_hi", *s.Statistics.TotalBytesInHi,
		"total_bytes_in_lo", *s.Statistics.TotalBytesInLo,
		"total_bytes_out", *s.Statistics.TotalBytesOut,
		"total_bytes_out_hi", *s.Statistics.TotalBytesOutHi,
		"total_bytes_out_lo", *s.Statistics.TotalBytesOutLo,
		"total_conn", *s.Statistics.TotalConn,
		"total_current_conn", *s.Statistics.TotalCurrentConn,
		"total_dns_responses", *s.Statistics.TotalDnsResponses,
		"total_requests", *s.Statistics.TotalRequests,
		"total_transactions", *s.Statistics.TotalTransactions,
		"up_time", *s.Statistics.UpTime,
	)
}

func vsStats(v *vtm.VirtualTrafficManager, vs string) {
	s, err := v.GetVirtualServerStatistics(vs)
	if err != nil {
		if err.ErrorId != "resource.not_found" {
			// silence errors for disabled vs
			log.Printf("ERROR: %+v\n", err)
		}
		return
	}

	fmt.Printf(
		"vtm_vs,%s=%s %s=%di,%s=%di,%s=%di,%s=%di,%s=%di,%s=%di,%s=%di,%s=%di,%s=%di,%s=%di,%s=%di,%s=%di,%s=%di,%s=%di,%s=%di,%s=%di,%s=%di,%s=%di,%s=%di,%s=%di,%s=%di,%s=%di,%s=%di,%s=%di,%s=%di,%s=%di,%s=%di,%s=%di,%s=%di,%s=\"%s\",%s=%di,%s=%di,%s=%di,%s=%di,%s=%di,%s=%di,%s=%di,%s=%di,%s=%di,%s=%di,%s=%di,%s=%di,%s=%di,%s=%di,%s=%di,%s=%di,%s=%di,%s=%di\n",
		"vs",
		escape(vs),
		"bytes_in", *s.Statistics.BytesIn,
		"bytes_in_hi", *s.Statistics.BytesInHi,
		"bytes_in_lo", *s.Statistics.BytesInLo,
		"bytes_out", *s.Statistics.BytesOut,
		"bytes_out_hi", *s.Statistics.BytesOutHi,
		"bytes_out_lo", *s.Statistics.BytesOutLo,
		"cert_status_requests", *s.Statistics.CertStatusRequests,
		"cert_status_responses", *s.Statistics.CertStatusResponses,
		"connect_timed_out", *s.Statistics.ConnectTimedOut,
		"connection_errors", *s.Statistics.ConnectionErrors,
		"connection_failures", *s.Statistics.ConnectionFailures,
		"current_conn", *s.Statistics.CurrentConn,
		"data_timed_out", *s.Statistics.DataTimedOut,
		"direct_replies", *s.Statistics.DirectReplies,
		"discard", *s.Statistics.Discard,
		"gzip", *s.Statistics.Gzip,
		"gzip_bytes_saved", *s.Statistics.GzipBytesSaved,
		"gzip_bytes_saved_hi", *s.Statistics.GzipBytesSavedHi,
		"gzip_bytes_saved_lo", *s.Statistics.GzipBytesSavedLo,
		"http_cache_hit_rate", *s.Statistics.HttpCacheHitRate,
		"http_cache_hits", *s.Statistics.HttpCacheHits,
		"http_cache_lookups", *s.Statistics.HttpCacheLookups,
		"http_rewrite_cookie", *s.Statistics.HttpRewriteCookie,
		"http_rewrite_location", *s.Statistics.HttpRewriteLocation,
		"keepalive_timed_out", *s.Statistics.KeepaliveTimedOut,
		"max_conn", *s.Statistics.MaxConn,
		"max_duration_timed_out", *s.Statistics.MaxDurationTimedOut,
		//"pkts_in", *s.Statistics.PktsIn,
		//"pkts_in_hi", *s.Statistics.PktsInHi,
		//"pkts_in_lo", *s.Statistics.PktsInLo,
		//"pkts_out", *s.Statistics.PktsOut,
		//"pkts_out_hi", *s.Statistics.PktsOutHi,
		//"pkts_out_lo", *s.Statistics.PktsOutLo,
		"port", *s.Statistics.Port,
		"processing_timed_out", *s.Statistics.ProcessingTimedOut,
		"protocol", stringFieldEscape(*s.Statistics.Protocol),
		"sip_rejected_requests", *s.Statistics.SipRejectedRequests,
		"sip_total_calls", *s.Statistics.SipTotalCalls,
		//"total_conn", *s.Statistics.TotalConn,
		"total_dgram", *s.Statistics.TotalDgram,
		"total_http1_requests", *s.Statistics.TotalHttp1Requests,
		"total_http1_requests_hi", *s.Statistics.TotalHttp1RequestsHi,
		"total_http1_requests_lo", *s.Statistics.TotalHttp1RequestsLo,
		"total_http2_requests", *s.Statistics.TotalHttp2Requests,
		"total_http2_requests_hi", *s.Statistics.TotalHttp2RequestsHi,
		"total_http2_requests_lo", *s.Statistics.TotalHttp2RequestsLo,
		"total_http_requests", *s.Statistics.TotalHttpRequests,
		"total_http_requests_hi", *s.Statistics.TotalHttpRequestsHi,
		"total_http_requests_lo", *s.Statistics.TotalHttpRequestsLo,
		"total_requests", *s.Statistics.TotalRequests,
		"total_requests_hi", *s.Statistics.TotalRequestsHi,
		"total_requests_lo", *s.Statistics.TotalRequestsLo,
		"total_tcp_reset", *s.Statistics.TotalTcpReset,
		"total_udp_unreachables", *s.Statistics.TotalUdpUnreachables,
		"udp_timed_out", *s.Statistics.UdpTimedOut,
	)
}

func slmStats(v *vtm.VirtualTrafficManager, slm string) {
	s, err := v.GetServiceLevelMonitorStatistics(slm)
	if err != nil {
		log.Printf("ERROR: %+v\n", err)
		return
	}

	fmt.Printf(
		"vtm_slm,%s=%s %s=%di,%s=%di,%s=\"%s\",%s=%di,%s=%di,%s=%di,%s=%di,%s=%di\n",
		"slm",
		escape(slm),
		"conforming",
		*s.Statistics.Conforming,
		"current_conns",
		*s.Statistics.CurrentConns,
		"is_ok",
		stringFieldEscape(*s.Statistics.IsOK),
		"response_max",
		*s.Statistics.ResponseMax,
		"response_mean",
		*s.Statistics.ResponseMean,
		"response_min",
		*s.Statistics.ResponseMin,
		"total_conn",
		*s.Statistics.TotalConn,
		"total_non_conf",
		*s.Statistics.TotalNonConf,
	)
}

func rateStats(v *vtm.VirtualTrafficManager, rate string) {
	s, err := v.GetConnectionRateLimitStatistics(rate)
	if err != nil {
		log.Printf("ERROR: %+v\n", err)
	}

	fmt.Printf(
		"vtm_rate,%s=%s %s=%di,%s=%di,%s=%di,%s=%di,%s=%di,%s=%di,%s=%di\n",
		"rate",
		escape(rate),
		"conns_entered",
		*s.Statistics.ConnsEntered,
		"conns_left",
		*s.Statistics.ConnsLeft,
		"current_rate",
		*s.Statistics.CurrentRate,
		"dropped",
		*s.Statistics.Dropped,
		"max_rate_per_min",
		*s.Statistics.MaxRatePerMin,
		"max_rate_per_sec",
		*s.Statistics.MaxRatePerSec,
		"queue_length",
		*s.Statistics.QueueLength,
	)
}

func poolStats(v *vtm.VirtualTrafficManager, pool string) {
	s, err := v.GetPoolStatistics(pool)
	if err != nil {
		log.Printf("ERROR: %+v\n", err)
		return
	}

	//
	fmt.Printf(
		"vtm_pool,%s=%s %s=\"%s\",%s=%di,%s=%di,%s=%di,%s=%di,%s=%di,%s=%di,%s=%di,%s=%di,%s=%di,%s=%di,%s=%di,%s=%di,%s=%di,%s=\"%s\",%s=%di,%s=%di,%s=\"%s\",%s=%di\n",
		"pool",
		escape(pool),
		"algorithm", escape(*s.Statistics.Algorithm),
		"bytes_in", *s.Statistics.BytesIn,
		"bytes_in_hi", *s.Statistics.BytesInHi,
		"bytes_in_lo", *s.Statistics.BytesInLo,
		"bytes_out", *s.Statistics.BytesOut,
		"bytes_out_hi", *s.Statistics.BytesOutHi,
		"bytes_out_lo", *s.Statistics.BytesOutLo,
		"conns_queued", *s.Statistics.ConnsQueued,
		"disabled", *s.Statistics.Disabled,
		"draining", *s.Statistics.Draining,
		"max_queue_time", *s.Statistics.MaxQueueTime,
		"mean_queue_time", *s.Statistics.MeanQueueTime,
		"min_queue_time", *s.Statistics.MinQueueTime,
		"nodes", *s.Statistics.Nodes,
		"persistence", stringFieldEscape(*s.Statistics.Persistence),
		"queue_timeouts", *s.Statistics.QueueTimeouts,
		"session_migrated", *s.Statistics.SessionMigrated,
		"state", stringFieldEscape(*s.Statistics.State),
		"total_conn", *s.Statistics.TotalConn,
	)
}

func nodeStats(v *vtm.VirtualTrafficManager, node string) {
	s, err := v.GetNodesNodeStatistics(node)
	if err != nil {
		if err.ErrorId != "statistics.no_counters" {
			// silence errors for disabled nodes
			log.Printf("ERROR: %+v\n", err)
		}
		return
	}

	//
	fmt.Printf(
		"vtm_node,%s=%s %s=%di,%s=%di,%s=%di,%s=%di,%s=%di,%s=%di,%s=%di,%s=%di,%s=%di,%s=%di,%s=%di,%s=%di,%s=%di,%s=%di,%s=\"%s\",%s=%di\n",
		"node",
		escape(node),
		"bytes_from_node_hi", *s.Statistics.BytesFromNodeHi,
		"bytes_from_node_lo", *s.Statistics.BytesFromNodeLo,
		"bytes_to_node_hi", *s.Statistics.BytesToNodeHi,
		"bytes_to_node_lo", *s.Statistics.BytesToNodeLo,
		"current_conn", *s.Statistics.CurrentConn,
		"current_requests", *s.Statistics.CurrentRequests,
		"errors", *s.Statistics.Errors,
		"failures", *s.Statistics.Failures,
		"new_conn", *s.Statistics.NewConn,
		"pooled_conn", *s.Statistics.PooledConn,
		"port", *s.Statistics.Port,
		"response_max", *s.Statistics.ResponseMax,
		"response_mean", *s.Statistics.ResponseMean,
		"response_min", *s.Statistics.ResponseMin,
		"state", stringFieldEscape(*s.Statistics.State),
		"total_conn", *s.Statistics.TotalConn,
	)
}

const (
	escapes            = "\t\n\f\r ,="
	nameEscapes        = "\t\n\f\r ,"
	stringFieldEscapes = "\t\n\f\r\\\""
)

// sourced from upstream: https://github.com/influxdata/telegraf/blob/master/plugins/serializers/influx/escape.go

var (
	escaper = strings.NewReplacer(
		"\t", `\t`,
		"\n", `\n`,
		"\f", `\f`,
		"\r", `\r`,
		`,`, `\,`,
		` `, `\ `,
		`=`, `\=`,
	)

	nameEscaper = strings.NewReplacer(
		"\t", `\t`,
		"\n", `\n`,
		"\f", `\f`,
		"\r", `\r`,
		`,`, `\,`,
		` `, `\ `,
	)

	stringFieldEscaper = strings.NewReplacer(
		"\t", `\t`,
		"\n", `\n`,
		"\f", `\f`,
		"\r", `\r`,
		`"`, `\"`,
		`\`, `\\`,
	)
)

// Escape a tagkey, tagvalue, or fieldkey
func escape(s string) string {
	if strings.ContainsAny(s, escapes) {
		return escaper.Replace(s)
	}

	return s
}

// Escape a measurement name
func nameEscape(s string) string {
	if strings.ContainsAny(s, nameEscapes) {
		return nameEscaper.Replace(s)
	}

	return s
}

// Escape a string field
func stringFieldEscape(s string) string {
	if strings.ContainsAny(s, stringFieldEscapes) {
		return stringFieldEscaper.Replace(s)
	}

	return s
}
