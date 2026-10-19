package cloud.eka_dev.eka_hub

import android.app.Application
import cloud.eka_dev.eka_hub.core.datastore.TokenManager
import cloud.eka_dev.eka_hub.core.network.ApiClient
import cloud.eka_dev.eka_hub.core.network.SSEClient

class EkaHubApp : Application() {

    lateinit var tokenManager: TokenManager
        private set

    lateinit var apiClient: ApiClient
        private set

    lateinit var sseClient: SSEClient
        private set

    override fun onCreate() {
        super.onCreate()
        tokenManager = TokenManager(this)
        apiClient = ApiClient(tokenManager)
        sseClient = SSEClient(tokenManager)
    }
}
