package com.example.security;

import java.util.List;
import org.springframework.stereotype.Service;

@Service
public class AuthorizationService {
  public boolean canAccess(String user, List<String> roles) {
    return roles.contains("admin");
  }
}
