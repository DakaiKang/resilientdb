#include <aws/lambda-runtime/runtime.h>
#include <aws/core/utils/json/JsonSerializer.h>
//#include <aws/lambda/LambdaClient.h>
#include <bits/stdint-uintn.h>
#include <set>
#include <xed25519.h>
#include <filters.h>
#include <regex>
#include <secblock.h>
#include <cmac.h>
#include <aes.h>
#include <cpr/cpr.h>

// change this to reflect the value of 'f' in the shim:
#define F 1
const std::string VERIFIER_HOST = "http://129.159.37.193";

typedef unsigned char byte;

using namespace aws::lambda_runtime;

bool ED25519verifyString(const std::string message, const std::string signature, std::string pkey)
{
    bool valid = true;
    byte byteKey[CryptoPP::ed25519PrivateKey::PUBLIC_KEYLENGTH];
    for (uint64_t i = 0; i < pkey.size(); i++)
    {
        byteKey[i] = pkey[i];
    }
    CryptoPP::ed25519::Verifier verifier(byteKey);
    CryptoPP::StringSource ss(signature + message, true, new CryptoPP::SignatureVerificationFilter(verifier, new CryptoPP::ArraySink((byte *)&valid, sizeof(valid))));
    return valid;
}

bool CMACverifyString(const std::string &publicKey, const std::string &message, const std::string &mac)
{
    bool valid = true;
    CryptoPP::SecByteBlock privKey((const unsigned char *)(publicKey.data()), publicKey.size());
    CryptoPP::CMAC<CryptoPP::AES> cmac(privKey.data(), privKey.size());
    const int flags = CryptoPP::HashVerificationFilter::THROW_EXCEPTION | CryptoPP::HashVerificationFilter::HASH_AT_END;
    CryptoPP::StringSource ss(message + mac, true, new CryptoPP::HashVerificationFilter(cmac, NULL, flags));
    return valid;
}

std::string decodeString(std::string encodedString)
{
    const std::regex comma(",");
    std::vector<std::string> byteVector{std::sregex_token_iterator(encodedString.begin() + 1, encodedString.end() - 1, comma, -1), std::sregex_token_iterator()};
    std::string decodedString = "";
    for (auto &s : byteVector)
    {
        char c = std::stoi(s);
        decodedString += c;
    }
    return decodedString;
}

bool verifyString(std::string option, std::string publicKey, std::string message, std::string signature, std::string &log)
{
    if (option == "CMAC")
    {
        log += " using CMAC ";
        return CMACverifyString(publicKey, message, signature);
    }
    else if (option == "ED25519")
    {
        log += " using ED25519 ";
        return ED25519verifyString(message, signature, publicKey);
    }
    else
    {
        log += " no valid method specified: " + option;
        return false;
    }
}

bool verifyCertificate(std::string certificate, std::string cryptoMethod, std::string &log)
{
    using namespace Aws::Utils::Json;
    bool valid = false;
    int numCommitMsgs = 0;
    std::set<std::string> keys;
    JsonValue certJson(certificate);
    assert(certJson.WasParseSuccessful());
    auto certView = certJson.View();
    Aws::Map<Aws::String, JsonView> certMap = certView.GetAllObjects();
    for (int i = 0; i < (2 * F + 1); i++)
    {
        std::string sigString, keyString, msgString;
        if (certView.ValueExists("signature" + std::to_string(i)))
        {
            sigString = decodeString(certMap["signature" + std::to_string(i)].WriteCompact());
        }
        if (certView.ValueExists("pkey" + std::to_string(i)))
        {
            keyString = decodeString(certMap["pkey" + std::to_string(i)].WriteCompact());
            keys.insert(keyString);
        }
        if (certView.ValueExists("message" + std::to_string(i)))
        {
            msgString = decodeString(certMap["message" + std::to_string(i)].WriteCompact());
        }
        if (sigString.size() != 0 && keyString.size() != 0 && msgString.size() != 0)
        {
            log += " obtained commit message: " + std::to_string(i);
            numCommitMsgs++;
            valid = verifyString(cryptoMethod, keyString, msgString, sigString, log);
        }
    }
    // For tests:
    //bool majoritySupport = (numCommitMsgs >= 1) && (keys.size() >= 1);
    // For normal case:
    bool majoritySupport = (numCommitMsgs >= 2 * F + 1) && (keys.size() >= (2 * F + 1));
    return valid && majoritySupport;
}

bool isWellFormedWriteRequest(Aws::Utils::Json::JsonView *requestView)
{
    return requestView->ValueExists("contracts") &&
           requestView->ValueExists("sequenceNumber") &&
           requestView->ValueExists("clientTs") &&
           requestView->ValueExists("uuid");
}

bool isWellFormedReadRequest(Aws::Utils::Json::JsonView *requestView)
{
    return requestView->ValueExists("sequenceNumber") && requestView->ValueExists("readSet");
}

bool isYCSB(Aws::Utils::Json::JsonView *requestView)
{
    return requestView->GetString("meta") == "ycsb";
}

bool isSmartContract(Aws::Utils::Json::JsonView *requestView)
{
    return requestView->GetString("meta") == "sc";
}

void sendWriteToVerifier(std::string jsonString, Aws::Utils::Json::JsonValue *response)
{
    cpr::Response r = cpr::Post(cpr::Url{VERIFIER_HOST + "/write"},
                                cpr::Header{{"Content-Type", "application/json"}},
                                cpr::Header{{"Connection", "close"}}, cpr::Body{jsonString},
                                cpr::ConnectTimeout{2000}, cpr::Timeout{2000});
    std::string text = r.text;
    if (r.status_code == 0)
    {
        text += r.error.message;
    }
    (*response).WithInteger("veriferStatusCode", r.status_code).WithString("verifierResponse", text);//.WithString("verifier_data", jsonString)
}

void sendReadToVerifier(std::string jsonString, Aws::Utils::Json::JsonValue *response)
{
    (*response).WithString("readReq:", jsonString);
    cpr::Response r = cpr::Post(cpr::Url{VERIFIER_HOST + "/read"},
                                cpr::Header{{"Content-Type", "application/json"}}, cpr::Body{jsonString},
                                cpr::ConnectTimeout{3000}, cpr::Timeout{3000});
    Aws::Utils::Json::JsonValue vresponse(r.text);
    (*response).WithInteger("veriferStatusCode", r.status_code).WithObject("snapshot", vresponse);
}

std::string getReadSet(Aws::Map<Aws::String, Aws::Utils::Json::JsonView> *smartContracts)
{
    std::string reads = "\"readSet\": [";
    Aws::Map<Aws::String, Aws::Utils::Json::JsonView>::iterator it = smartContracts->begin();
    // Need to do this due to a lack of operator support:
    Aws::Map<Aws::String, Aws::Utils::Json::JsonView>::iterator secondToLast = smartContracts->end();
    secondToLast--;
    while (it != smartContracts->end())
    {
        Aws::Utils::Json::JsonValue contract(it->second.WriteCompact());
        if (it != secondToLast)
        {
            reads += "\"" + contract.View().GetString("dest") + "\",";
        }
        else
        {
            reads += "\"" + contract.View().GetString("dest") + "\"";
        }
        it++;
    }
    reads += "]";
    return reads;
}

std::string getWriteSet(Aws::Map<Aws::String, Aws::Utils::Json::JsonView> *smartContracts, Aws::Utils::Json::JsonView *snapshot)
{
    using namespace Aws::Utils::Json;
    std::string contractSet = "\"contracts\":{";
    Aws::Map<Aws::String, JsonView>::iterator it = smartContracts->begin();
    Aws::Map<Aws::String, JsonView>::iterator secondToLast = smartContracts->end();
    secondToLast--;
    while (it != smartContracts->end())
    {
        JsonValue contract(it->second.WriteCompact());

        int type = contract.View().GetInteger("type");

        switch (type)
        {
        case 0:
        {
            std::string srcKey = contract.View().GetString("src");
            std::string destKey = contract.View().GetString("dest");
            uint64_t dest = 0;
            uint64_t src = 0;
            if (snapshot->KeyExists(srcKey) && snapshot->KeyExists(destKey))
            {
                dest = std::stoull(snapshot->GetString(destKey));
                src = std::stoull(snapshot->GetString(srcKey));
            }
            uint64_t amt = std::stoi(contract.View().GetString("amt"));
            contractSet += "\"" + srcKey + "\":\"" + std::to_string(src - amt) + "\",";
            contractSet += "\"" + destKey + "\":\"" + std::to_string(dest + amt) + "\"";
            break;
        }
        case 1:
        {
            std::string key = contract.View().GetString("dest");
            uint64_t dest = 0;
            if (snapshot->KeyExists(key))
            {
                dest = std::stoull(snapshot->GetString(key));
            }
            uint64_t amt = std::stoi(contract.View().GetString("amt"));
            contractSet += "\"" + key + "\":\"" + std::to_string(dest + amt) + "\"";
            break;
        }
        case 2:
        {
            std::string key = contract.View().GetString("src");
            uint64_t src = 0;
            if (snapshot->KeyExists(key))
            {
                src = std::stoull(snapshot->GetString(key));
            }
            uint64_t amt = std::stoi(contract.View().GetString("amt"));
            contractSet += "\"" + key + "\":\"" + std::to_string(src - amt) + "\"";
            break;
        }
        }
        if (it != secondToLast)
        {
            contractSet += ",";
        }
        it++;
    }

    contractSet += "}";
    return contractSet;
}

void executeSmartContracts(Aws::Utils::Json::JsonView *requestView, Aws::Utils::Json::JsonValue *response)
{
    using namespace Aws::Utils::Json;
    auto smartContractsView = requestView->GetObject("contracts");
    auto smartContracts = smartContractsView.GetAllObjects();
    /* Here the naming conventions are a bit confusing. 'ReadSet' has a slightly
   * different meaning at different stages of the protocol. Below, 'ReadSet'
   * refers to the set of keys that we want to read.
   * e.g.:[k0, k1, k2, ..... kn]
   */
    std::string readReq = "{\"sequenceNumber\":\"" + requestView->GetString("sequenceNumber") + "\"," +
                          getReadSet(&smartContracts) + "}";
    sendReadToVerifier(readReq, response);
    /* Once we've obtained the read set we can proceed with computing the write
   * set. Note that this is where the definition of 'ReadSet' changes. Now, the
   * 'ReadSet' or 'snapshot' refers to the set of keys that we wanted to read,
   * and the values associated with those keys (i.e. the values read).
   * e.g.:{k0:v0, k1:v1, k2:v2, ..... kn:vn}
   */
    JsonView snapshot = (response->View()).GetObject("snapshot");
    std::string writeReq = "{\"sequenceNumber\":\"" + requestView->GetString("sequenceNumber") + "\"," +
                           "\"uuid\":\"" + requestView->GetString("uuid") + "\"," +
                           "\"clientId\":\"" + requestView->GetString("clientId") + "\"," +
                           "\"readSet\":" + snapshot.WriteCompact() + "," +
                           getWriteSet(&smartContracts, &snapshot) + "}";
    sendWriteToVerifier(writeReq, response);
}

invocation_response lambda_handler(invocation_request const &request)
{
    using namespace Aws::Utils::Json;

    std::string log;
    JsonValue response;
    bool certified = false;

    // Obtain View:
    JsonValue requestJson(request.payload);
    assert(requestJson.WasParseSuccessful());
    auto requestView = requestJson.View();

    // Convert View into map in order to more easily access items:
    Aws::Map<Aws::String, JsonView> payloadMap = requestView.GetAllObjects();

    // We first check to see if the request has the certificate:
    if (requestView.ValueExists("certificate") && requestView.ValueExists("cryptoMethod") && requestView.ValueExists("meta"))
    {
        // Extract the certificate and attempt to verify it:
        std::string certificate = payloadMap["certificate"].WriteCompact();
        std::string cryptoMethod = requestView.GetString("cryptoMethod");
        certified = verifyCertificate(certificate, cryptoMethod, log);
        if (certified)
        {
            log += " verified certificate " + requestView.GetString("meta") + " ";
        }
    }
    else
    {
        response.WithString("error", "missing certificate/meta").WithString("executorLog", log);
        auto const apig_response = response.View().WriteCompact();
        return aws::lambda_runtime::invocation_response::success(apig_response, "application/json");
    }
    /* If the request's certificate can be verified and if the request is a well-formed write request, we transmit the
   * writes to the verifier:
   */
    if (certified && isWellFormedWriteRequest(&requestView))
    {
        if (isYCSB(&requestView))
        {
            requestJson.WithObject("certificate", JsonValue());
            sendWriteToVerifier(requestJson.View().WriteCompact(), &response);
            response.WithString("executorLog", log);
            auto const apig_response = response.View().WriteCompact();
            return aws::lambda_runtime::invocation_response::success(apig_response, "application/json");
        }
        else if (isSmartContract(&requestView))
        {
            executeSmartContracts(&requestView, &response);
            response.WithString("executorLog", log);
            auto const apig_response = response.View().WriteCompact();
            return aws::lambda_runtime::invocation_response::success(apig_response, "application/json");
        }
        else
        {
            response.WithString("error", "non-supported txn type").WithString("executorLog", log);
            auto const apig_response = response.View().WriteCompact();
            return aws::lambda_runtime::invocation_response::success(apig_response, "application/json");
        }
    }
    else if (certified && isWellFormedReadRequest(&requestView))
    {
        sendReadToVerifier(requestView.WriteCompact(), &response);
        response.WithString("executorLog", log);
        auto const apig_response = response.View().WriteCompact();
        return aws::lambda_runtime::invocation_response::success(apig_response, "application/json");
    }
    else
    {
        std::string errorString;
        if (!certified)
        {
            errorString = "certification failed";
        }
        else
        {
            errorString = "malformed request";
        }
        response.WithString("error", errorString).WithString("executorLog", log);
        auto const apig_response = response.View().WriteCompact();
        return aws::lambda_runtime::invocation_response::success(apig_response, "application/json");
    }
}

int main()
{
    run_handler(lambda_handler);
    return 0;
}
