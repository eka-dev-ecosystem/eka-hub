package cloud.eka_dev.eka_hub.core.network

import com.jakewharton.retrofit2.converter.kotlinx.serialization.asConverterFactory
import cloud.eka_dev.eka_hub.core.datastore.TokenManager
import cloud.eka_dev.eka_hub.core.model.*
import kotlinx.coroutines.runBlocking
import kotlinx.serialization.json.Json
import okhttp3.*
import okhttp3.MediaType.Companion.toMediaType
import okhttp3.logging.HttpLoggingInterceptor
import retrofit2.Retrofit
import retrofit2.http.*
import java.util.concurrent.TimeUnit

interface HubApiService {
    @GET("hub/briefing")
    suspend fun getDailyBriefing(): BaseResponse<DailyBriefingResponse>

    @GET("hub/summaries")
    suspend fun getAppSummaries(): BaseResponse<List<AppSummaryResponse>>

    @POST("hub/chat")
    suspend fun sendChatMessage(@Body req: Map<String, String>): BaseResponse<ChatMessage>

    @POST("hub/link/initiate")
    suspend fun initiateLink(@Body req: Map<String, String>): BaseResponse<Map<String, String>>
}

class ApiClient(private val tokenManager: TokenManager) {

    private val json = Json {
        ignoreUnknownKeys = true
        isLenient = true
    }

    private val okHttpClient: OkHttpClient = OkHttpClient.Builder()
        .connectTimeout(15, TimeUnit.SECONDS)
        .readTimeout(15, TimeUnit.SECONDS)
        .addInterceptor { chain ->
            val requestBuilder = chain.request().newBuilder()
            runBlocking {
                val token = tokenManager.getAccessToken()
                if (!token.isNullOrEmpty()) {
                    requestBuilder.addHeader("Authorization", "Bearer $token")
                }
            }
            chain.proceed(requestBuilder.build())
        }
        .addInterceptor(HttpLoggingInterceptor().apply {
            level = HttpLoggingInterceptor.Level.BODY
        })
        .build()

    private val retrofit: Retrofit = Retrofit.Builder()
        .baseUrl("https://be-hub.eka-dev.cloud/api/v1/")
        .client(okHttpClient)
        .addConverterFactory(json.asConverterFactory("application/json".toMediaType()))
        .build()

    val apiService: HubApiService = retrofit.create(HubApiService::class.java)
}
