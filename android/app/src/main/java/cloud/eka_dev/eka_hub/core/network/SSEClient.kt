package cloud.eka_dev.eka_hub.core.network

import cloud.eka_dev.eka_hub.core.datastore.TokenManager
import kotlinx.coroutines.channels.awaitClose
import kotlinx.coroutines.flow.Flow
import kotlinx.coroutines.flow.callbackFlow
import okhttp3.*
import okhttp3.sse.EventSource
import okhttp3.sse.EventSourceListener
import okhttp3.sse.EventSources
import java.util.concurrent.TimeUnit

class SSEClient(private val tokenManager: TokenManager) {

    private val client = OkHttpClient.Builder()
        .readTimeout(0, TimeUnit.MILLISECONDS)
        .build()

    fun listenToHubEvents(): Flow<String> = callbackFlow {
        val request = Request.Builder()
            .url("https://be-hub.eka-dev.cloud/api/v1/realtime/sse/events")
            .build()

        val listener = object : EventSourceListener() {
            override fun onEvent(eventSource: EventSource, id: String?, type: String?, data: String) {
                trySend(data)
            }

            override fun onFailure(eventSource: EventSource, t: Throwable?, response: Response?) {
                // Handle reconnection gracefully
            }
        }

        val eventSource = EventSources.createFactory(client).newEventSource(request, listener)
        awaitClose {
            eventSource.cancel()
        }
    }
}
