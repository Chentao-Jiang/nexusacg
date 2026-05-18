import 'package:flutter/material.dart';
import 'package:nexusacg/app.dart';
import 'package:nexusacg/core/network/api_client.dart';

void main() async {
  WidgetsFlutterBinding.ensureInitialized();
  await ApiClient().init();
  runApp(const NexusACGApp());
}
