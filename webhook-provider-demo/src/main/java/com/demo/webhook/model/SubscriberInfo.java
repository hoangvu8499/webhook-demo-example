package com.demo.webhook.model;

import com.fasterxml.jackson.annotation.JsonIgnoreProperties;
import com.fasterxml.jackson.annotation.JsonProperty;

import java.util.List;
import java.util.Map;

@JsonIgnoreProperties(ignoreUnknown = true)
public class SubscriberInfo {

    private String id;
    private String status;
    private String email;
    private String source;

    @JsonProperty("first_name")
    private String firstName;

    @JsonProperty("last_name")
    private String lastName;

    private List<Map<String, Object>> segments;

    @JsonProperty("custom_fields")
    private Map<String, String> customFields;

    @JsonProperty("optin_ip")
    private String optinIp;

    @JsonProperty("optin_timestamp")
    private String optinTimestamp;

    @JsonProperty("created_at")
    private String createdAt;

    public String getId() { return id; }
    public void setId(String id) { this.id = id; }

    public String getStatus() { return status; }
    public void setStatus(String status) { this.status = status; }

    public String getEmail() { return email; }
    public void setEmail(String email) { this.email = email; }

    public String getSource() { return source; }
    public void setSource(String source) { this.source = source; }

    public String getFirstName() { return firstName; }
    public void setFirstName(String firstName) { this.firstName = firstName; }

    public String getLastName() { return lastName; }
    public void setLastName(String lastName) { this.lastName = lastName; }

    public List<Map<String, Object>> getSegments() { return segments; }
    public void setSegments(List<Map<String, Object>> segments) { this.segments = segments; }

    public Map<String, String> getCustomFields() { return customFields; }
    public void setCustomFields(Map<String, String> customFields) { this.customFields = customFields; }

    public String getOptinIp() { return optinIp; }
    public void setOptinIp(String optinIp) { this.optinIp = optinIp; }

    public String getOptinTimestamp() { return optinTimestamp; }
    public void setOptinTimestamp(String optinTimestamp) { this.optinTimestamp = optinTimestamp; }

    public String getCreatedAt() { return createdAt; }
    public void setCreatedAt(String createdAt) { this.createdAt = createdAt; }
}
