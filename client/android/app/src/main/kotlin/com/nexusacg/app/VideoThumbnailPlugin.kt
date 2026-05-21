package com.nexusacg.app

import android.graphics.Bitmap
import android.media.MediaMetadataRetriever
import io.flutter.embedding.engine.FlutterEngine
import io.flutter.plugin.common.MethodChannel
import java.io.File
import java.io.FileOutputStream

class VideoThumbnailPlugin(private val flutterEngine: FlutterEngine) {
    init {
        MethodChannel(flutterEngine.dartExecutor.binaryMessenger, "nexusacg/video_thumbnail").setMethodCallHandler { call, result ->
            if (call.method == "getThumbnail") {
                val path = call.argument<String>("path") ?: ""
                try {
                    val retriever = MediaMetadataRetriever()
                    retriever.setDataSource(path)
                    val bitmap = retriever.frameAtTime(1000000)
                    retriever.release()
                    if (bitmap != null) {
                        val cacheDir = File(flutterEngine.dartExecutor.binaryMessenger.toString().hashCode().let {
                            "/data/data/com.nexusacg.app/cache"
                        })
                        val thumbFile = File("/data/data/com.nexusacg.app/cache", "thumb_${System.currentTimeMillis()}.jpg")
                        thumbFile.parentFile?.mkdirs()
                        FileOutputStream(thumbFile).use { fos ->
                            bitmap.compress(Bitmap.CompressFormat.JPEG, 80, fos)
                        }
                        bitmap.recycle()
                        result.success(thumbFile.absolutePath)
                    } else {
                        result.success(null)
                    }
                } catch (e: Exception) {
                    result.success(null)
                }
            } else {
                result.notImplemented()
            }
        }
    }
}
