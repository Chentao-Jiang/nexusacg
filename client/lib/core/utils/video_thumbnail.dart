import 'dart:io';
import 'package:flutter/services.dart';

class VideoThumbnail {
  static const _channel = MethodChannel('nexusacg/video_thumbnail');

  static Future<File?> getThumbnail(String videoPath) async {
    try {
      final result = await _channel.invokeMethod('getThumbnail', {'path': videoPath});
      if (result is String && result.isNotEmpty) {
        final file = File(result);
        if (await file.exists()) return file;
      }
    } catch (_) {}
    return null;
  }
}
