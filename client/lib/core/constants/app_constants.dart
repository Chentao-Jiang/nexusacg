class AppConstants {
  static const String apiBaseUrl = String.fromEnvironment('API_BASE_URL', defaultValue: 'http://101.133.169.72:8080/api/v1');

  static const String zoneCosplay = 'cosplay';
  static const String zonePeripheral = 'peripheral';

  static const int defaultPageSize = 20;
}
